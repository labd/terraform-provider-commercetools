package connect_connector_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/connect"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/resources/connect_connector"
)

const projectKey = "test-project"

// fakeConnect holds connectors the way the Connect API does: a published one
// is Processing once, then Published (or Failed when its tag is "broken"),
// and each publish raises the version of the published connector.
type fakeConnect struct {
	*httptest.Server
	mu        sync.Mutex
	staged    map[string]*connect.ConnectorStaged
	published map[string]int
	publishes int
}

func newFakeConnect(t *testing.T) *fakeConnect {
	f := &fakeConnect{staged: map[string]*connect.ConnectorStaged{}, published: map[string]int{}}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("POST /connectors/drafts", f.locked(f.create))
	mux.HandleFunc("GET /connectors/drafts/{ref}", f.locked(f.get))
	mux.HandleFunc("POST /connectors/drafts/{ref}", f.locked(f.update))
	mux.HandleFunc("DELETE /connectors/drafts/{ref}", f.locked(f.delete))
	mux.HandleFunc("GET /connectors/{ref}", f.locked(f.getPublished))
	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

func (f *fakeConnect) providerConfig() string {
	return fmt.Sprintf(`
provider "commercetools" {
  client_id     = "client"
  client_secret = "secret"
  project_key   = %q
  scopes        = "manage_connectors:%s"
  api_url       = %q
  token_url     = %q
  connect_url   = %q
}
`, projectKey, projectKey, f.URL, f.URL, f.URL)
}

func (f *fakeConnect) locked(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			write(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		next(w, r)
	}
}

// lookup finds a staged connector by `key=<key>` or by id.
func (f *fakeConnect) lookup(ref string) *connect.ConnectorStaged {
	if key, ok := strings.CutPrefix(ref, "key="); ok {
		return f.staged[key]
	}
	for _, c := range f.staged {
		if c.ID == ref {
			return c
		}
	}
	return nil
}

func (f *fakeConnect) create(w http.ResponseWriter, r *http.Request) {
	var draft connect.ConnectorStagedDraft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	c := &connect.ConnectorStaged{
		ID:               "connector-" + *draft.Key,
		Key:              draft.Key,
		Version:          1,
		Name:             draft.Name,
		IntegrationTypes: draft.IntegrationTypes,
		Creator:          draft.Creator,
		Repository:       draft.Repository,
		Private:          true,
		PrivateProjects:  draft.PrivateProjects,
		SupportedRegions: draft.SupportedRegions,
		Status:           string(connect.ConnectorStatusDraft),
		HasChanges:       true,
		IsPreviewable:    "none",
		DocumentationUrl: draft.DocumentationUrl,
	}
	if draft.Description != nil {
		c.Description = *draft.Description
	}
	f.staged[*draft.Key] = c
	write(w, http.StatusCreated, c)
}

func (f *fakeConnect) get(w http.ResponseWriter, r *http.Request) {
	c := f.lookup(r.PathValue("ref"))
	if c == nil {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	if c.Status == string(connect.ConnectorStatusProcessing) {
		if c.Repository.Tag == "broken" {
			c.Status = string(connect.ConnectorStatusFailed)
			c.PublishingReport = &connect.ConnectorReport{Entries: []connect.ConnectorReportEntry{{
				Type: connect.ConnectorReportEntryTypeError, Title: "Build failed",
				Message: connectRef("tag broken has no connect.yaml"),
			}}}
		} else {
			c.Status = string(connect.ConnectorStatusPublished)
			c.HasChanges = false
			f.published[*c.Key]++
		}
	}
	write(w, http.StatusOK, c)
}

func (f *fakeConnect) getPublished(w http.ResponseWriter, r *http.Request) {
	c := f.lookup(r.PathValue("ref"))
	if c == nil || f.published[*c.Key] == 0 {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	write(w, http.StatusOK, connect.Connector{ID: c.ID, Key: c.Key, Version: f.published[*c.Key], Name: c.Name})
}

func (f *fakeConnect) update(w http.ResponseWriter, r *http.Request) {
	c := f.lookup(r.PathValue("ref"))
	if c == nil {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	var body struct {
		Version int               `json:"version"`
		Actions []json.RawMessage `json:"actions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	if body.Version != c.Version {
		write(w, http.StatusConflict, map[string]string{"message": "version mismatch"})
		return
	}
	for _, raw := range body.Actions {
		var action struct {
			Action           string                    `json:"action"`
			Name             string                    `json:"name"`
			Description      string                    `json:"description"`
			URL              string                    `json:"url"`
			Tag              string                    `json:"tag"`
			IntegrationTypes []connect.IntegrationType `json:"integrationTypes"`
			PrivateProject   string                    `json:"privateProject"`
			Certification    bool                      `json:"certification"`
		}
		_ = json.Unmarshal(raw, &action)
		switch action.Action {
		case "setName":
			c.Name = action.Name
		case "setDescription":
			c.Description = action.Description
		case "setRepository":
			c.Repository = connect.Repository{Url: action.URL, Tag: action.Tag}
		case "setIntegrationTypes":
			c.IntegrationTypes = action.IntegrationTypes
		case "addPrivateProject":
			c.PrivateProjects = append(c.PrivateProjects, action.PrivateProject)
		case "removePrivateProject":
			var kept []string
			for _, p := range c.PrivateProjects {
				if p != action.PrivateProject {
					kept = append(kept, p)
				}
			}
			c.PrivateProjects = kept
		case "publish":
			if action.Certification {
				write(w, http.StatusBadRequest, map[string]string{"message": "certification not expected"})
				return
			}
			f.publishes++
			c.Status = string(connect.ConnectorStatusProcessing)
		default:
			write(w, http.StatusBadRequest, map[string]string{"message": "unknown action " + action.Action})
			return
		}
		if action.Action != "publish" {
			c.HasChanges = true
		}
	}
	c.Version++
	write(w, http.StatusCreated, c)
}

func (f *fakeConnect) delete(w http.ResponseWriter, r *http.Request) {
	c := f.lookup(r.PathValue("ref"))
	if c == nil {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	delete(f.staged, *c.Key)
	delete(f.published, *c.Key)
	write(w, http.StatusOK, c)
}

// release does what the commercetools CLI does on a release: set a new tag
// and publish it.
func (f *fakeConnect) release(key, tag string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := f.staged[key]
	c.Repository.Tag = tag
	c.Version += 2
	c.HasChanges = false
	c.Status = string(connect.ConnectorStatusPublished)
	f.published[key]++
}

func (f *fakeConnect) connector(key string) (*connect.ConnectorStaged, int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.staged[key], f.published[key], f.publishes
}

func connectRef(s string) *string { return &s }

func write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func config(name, tag, projects string) string {
	return fmt.Sprintf(`
resource "commercetools_connect_connector" "sync" {
  key               = "evolve-sync"
  name              = %q
  repository_url    = "git@github.com:labd/evolve-sync.git"
  repository_tag    = %q
  integration_types = ["search"]
  supported_regions = ["europe-west1.gcp"]
  private_projects  = [%s]
  publish           = true

  creator {
    email   = "dev@example.com"
    company = "Lab Digital"
  }

  lifecycle {
    ignore_changes = [repository_tag]
  }
}
`, name, tag, projects)
}

func TestConnectConnector(t *testing.T) {
	defer connect_connector.SetPollInterval(time.Millisecond)()
	fake := newFakeConnect(t)
	name := "commercetools_connect_connector.sync"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.FrameworkProviderFactories(),
		CheckDestroy: func(_ *terraform.State) error {
			if c, _, _ := fake.connector("evolve-sync"); c != nil {
				return errors.New("the connector still exists after destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: fake.providerConfig() + `
resource "commercetools_connect_connector" "sync" {
  key              = "evolve-sync"
  name             = "Evolve sync"
  repository_url   = "git@github.com:labd/evolve-sync.git"
  repository_tag   = "v1.0.0"
  private_projects = ["test-project"]
  creator {
    email = "dev@example.com"
  }
}
`,
				ExpectError: regexp.MustCompile(`must be\s+\{region\}:\{projectKey\}`),
			},
			{
				Config: fake.providerConfig() + config("Evolve sync", "v1.0.0", `"europe-west1.gcp:test-project"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "connector-evolve-sync"),
					resource.TestCheckResourceAttr(name, "status", "Published"),
					resource.TestCheckResourceAttr(name, "has_changes", "false"),
					resource.TestCheckResourceAttr(name, "published_version", "1"),
					resource.TestCheckResourceAttr(name, "creator.company", "Lab Digital"),
					resource.TestCheckNoResourceAttr(name, "description"),
					func(_ *terraform.State) error {
						c, _, _ := fake.connector("evolve-sync")
						assert.Equal(t, "v1.0.0", c.Repository.Tag)
						assert.Equal(t, []string{"europe-west1.gcp:test-project"}, c.PrivateProjects)
						return nil
					},
				),
			},
			{
				// A release by the CLI is kept: the tag is ignored and the
				// published version follows it, without publishing again.
				PreConfig: func() { fake.release("evolve-sync", "v1.1.0") },
				Config:    fake.providerConfig() + config("Evolve sync", "v1.0.0", `"europe-west1.gcp:test-project"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "repository_tag", "v1.1.0"),
					resource.TestCheckResourceAttr(name, "published_version", "2"),
					func(_ *terraform.State) error {
						c, _, publishes := fake.connector("evolve-sync")
						assert.Equal(t, "v1.1.0", c.Repository.Tag)
						assert.Equal(t, 1, publishes)
						return nil
					},
				),
			},
			{
				// A change by Terraform is published.
				Config: fake.providerConfig() + config("Evolve Algolia sync", "v1.0.0",
					`"europe-west1.gcp:test-project", "europe-west1.gcp:other-project"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Evolve Algolia sync"),
					resource.TestCheckResourceAttr(name, "repository_tag", "v1.1.0"),
					resource.TestCheckResourceAttr(name, "published_version", "3"),
					resource.TestCheckResourceAttr(name, "private_projects.#", "2"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateId:           "evolve-sync",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestConnectConnectorPublishFailure(t *testing.T) {
	defer connect_connector.SetPollInterval(time.Millisecond)()
	fake := newFakeConnect(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.FrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      fake.providerConfig() + config("Evolve sync", "broken", `"europe-west1.gcp:test-project"`),
				ExpectError: regexp.MustCompile(`Build failed: tag broken\s+has no connect.yaml`),
			},
		},
	})
}
