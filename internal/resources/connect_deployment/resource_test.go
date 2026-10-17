package connect_deployment_test

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
	"github.com/labd/terraform-provider-commercetools/internal/resources/connect_deployment"
)

const projectKey = "test-project"

// latestConnectorVersion is the version a redeploy that updates the
// connector moves a deployment to; a new one is on version 3.
const latestConnectorVersion = 4

// fakeConnect holds deployments the way the Connect API does: a new or
// redeployed one is Queued, then Deploying, then Deployed (or Failed) on
// successive reads; a deleted one is Undeploying once, then gone. Secured
// values are kept but never returned.
type fakeConnect struct {
	*httptest.Server
	mu          sync.Mutex
	deployments map[string]*connect.Deployment
	secrets     map[string]map[string]string
	failNext    bool
	failing     map[string]bool
	steps       map[string]int
	deleted     map[string]bool
}

func newFakeConnect(t *testing.T) *fakeConnect {
	f := &fakeConnect{
		deployments: map[string]*connect.Deployment{},
		secrets:     map[string]map[string]string{},
		failing:     map[string]bool{},
		steps:       map[string]int{},
		deleted:     map[string]bool{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("POST /"+projectKey+"/deployments", f.locked(f.create))
	mux.HandleFunc("GET /"+projectKey+"/deployments/{ref}", f.locked(f.get))
	mux.HandleFunc("POST /"+projectKey+"/deployments/{ref}", f.locked(f.update))
	mux.HandleFunc("DELETE /"+projectKey+"/deployments/{ref}", f.locked(f.delete))
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
  scopes        = "manage_connectors_deployments:%s"
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

// lookup finds a deployment by `key=<key>` or by id.
func (f *fakeConnect) lookup(ref string) *connect.Deployment {
	if key, ok := strings.CutPrefix(ref, "key="); ok {
		return f.deployments[key]
	}
	for _, d := range f.deployments {
		if d.ID == ref {
			return d
		}
	}
	return nil
}

func (f *fakeConnect) create(w http.ResponseWriter, r *http.Request) {
	var draft connect.DeploymentDraft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	key := *draft.Connector.Key
	if key != "evolve-sync" {
		write(w, http.StatusBadRequest, map[string]any{
			"statusCode": 400, "message": "connector not found",
			"errors": []map[string]string{{"code": "InvalidInput", "message": "connector not found"}},
		})
		return
	}
	deploymentType := "sandbox"
	if draft.Type != nil {
		deploymentType = *draft.Type
	}
	d := &connect.Deployment{
		ID:             "deployment-" + *draft.Key,
		Key:            draft.Key,
		Version:        1,
		Type:           deploymentType,
		Connector:      connect.DeploymentConnector{ID: "connector-1", Key: &key, Version: 3},
		DeployedRegion: draft.Region,
		Status:         string(connect.DeploymentStatusQueued),
	}
	f.applyConfig(d, draft.Configurations, draft.GlobalConfiguration)
	f.deployments[*draft.Key] = d
	f.failing[*draft.Key] = f.failNext
	f.failNext = false
	f.steps[*draft.Key] = 0
	f.deleted[*draft.Key] = false
	write(w, http.StatusCreated, f.answer(d))
}

func (f *fakeConnect) applyConfig(d *connect.Deployment, apps []connect.DeploymentConfigurationApplication, global *connect.DeploymentGlobalConfiguration) {
	d.Applications = nil
	secrets := map[string]string{}
	for _, app := range apps {
		deployed := connect.ApplicationDeployment{
			ID:                    "app-" + app.ApplicationName,
			ApplicationName:       app.ApplicationName,
			StandardConfiguration: app.StandardConfiguration,
			Schedule:              app.Schedule,
		}
		if app.ApplicationName == "service" {
			deployed.Url = connectRef("https://service-" + *d.Key + ".europe-west1.gcp.commercetools.app")
		}
		if app.ApplicationName == "job" && app.Schedule == nil {
			// The schedule from connect.yaml, which the configuration need not repeat.
			deployed.Schedule = connectRef("0 * * * *")
		}
		for _, value := range app.SecuredConfiguration {
			secrets[app.ApplicationName+"/"+value.Key] = value.Value
		}
		d.Applications = append(d.Applications, deployed)
	}
	if global != nil {
		d.GlobalConfiguration = &connect.DeploymentGlobalConfiguration{StandardConfiguration: global.StandardConfiguration}
		for _, value := range global.SecuredConfiguration {
			secrets["global/"+value.Key] = value.Value
		}
	}
	f.secrets[*d.Key] = secrets
}

// answer is the deployment as the API returns it, without secured values.
func (f *fakeConnect) answer(d *connect.Deployment) *connect.Deployment {
	copied := *d
	copied.Applications = append([]connect.ApplicationDeployment{}, d.Applications...)
	for i := range copied.Applications {
		copied.Applications[i].SecuredConfiguration = []connect.ConfigurationValue{}
	}
	return &copied
}

func (f *fakeConnect) get(w http.ResponseWriter, r *http.Request) {
	d := f.lookup(r.PathValue("ref"))
	if d == nil {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	key := *d.Key
	if f.deleted[key] {
		delete(f.deployments, key)
		write(w, http.StatusOK, f.answer(d))
		return
	}
	f.steps[key]++
	switch {
	case d.Status == string(connect.DeploymentStatusQueued):
		d.Status = string(connect.DeploymentStatusDeploying)
	case d.Status == string(connect.DeploymentStatusDeploying) && f.failing[key]:
		d.Status = string(connect.DeploymentStatusFailed)
	case d.Status == string(connect.DeploymentStatusDeploying):
		d.Status = string(connect.DeploymentStatusDeployed)
	}
	write(w, http.StatusOK, f.answer(d))
}

func (f *fakeConnect) update(w http.ResponseWriter, r *http.Request) {
	d := f.lookup(r.PathValue("ref"))
	if d == nil {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	var body struct {
		Version int `json:"version"`
		Actions []struct {
			Action              string                                       `json:"action"`
			UpdateConnector     *bool                                        `json:"updateConnector"`
			ConfigurationValues []connect.DeploymentConfigurationApplication `json:"configurationValues"`
			GlobalConfiguration *connect.DeploymentGlobalConfiguration       `json:"globalConfiguration"`
		} `json:"actions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	if body.Version != d.Version {
		write(w, http.StatusConflict, map[string]string{"message": "version mismatch"})
		return
	}
	for _, action := range body.Actions {
		if action.Action != "redeploy" {
			write(w, http.StatusBadRequest, map[string]string{"message": "unknown action " + action.Action})
			return
		}
		f.applyConfig(d, action.ConfigurationValues, action.GlobalConfiguration)
		if action.UpdateConnector != nil && *action.UpdateConnector {
			d.Connector.Version = latestConnectorVersion
		}
	}
	d.Version++
	d.Status = string(connect.DeploymentStatusQueued)
	f.failing[*d.Key] = false
	write(w, http.StatusCreated, f.answer(d))
}

func (f *fakeConnect) delete(w http.ResponseWriter, r *http.Request) {
	d := f.lookup(r.PathValue("ref"))
	if d == nil {
		write(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	d.Status = string(connect.DeploymentStatusUndeploying)
	f.deleted[*d.Key] = true
	write(w, http.StatusAccepted, f.answer(d))
}

func (f *fakeConnect) deployment(key string) (*connect.Deployment, map[string]string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deployments[key], f.secrets[key]
}

func connectRef(s string) *string { return &s }

func write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func config(appKey, schedule string) string {
	return configVersion(appKey, schedule, "")
}

func configVersion(appKey, schedule, version string) string {
	return fmt.Sprintf(`
resource "commercetools_connect_deployment" "sync" {
  key           = "evolve-sync"
  connector_key = "evolve-sync"
  region        = "europe-west1.gcp"
  %s

  global_standard_configuration = {
    CTP_REGION = "europe-west1.gcp"
  }

  application {
    name = "service"
    standard_configuration = {
      ALGOLIA_APP_ID = "EXAMPLEAPP"
    }
    secured_configuration = {
      ALGOLIA_ADMIN_KEY = %q
    }
  }

  application {
    name = "job"
    %s
  }
}
`, version, appKey, schedule)
}

func TestConnectDeployment(t *testing.T) {
	defer connect_deployment.SetPollInterval(time.Millisecond)()
	fake := newFakeConnect(t)
	name := "commercetools_connect_deployment.sync"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.FrameworkProviderFactories(),
		CheckDestroy: func(_ *terraform.State) error {
			if d, _ := fake.deployment("evolve-sync"); d != nil {
				return errors.New("the deployment still exists after destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: fake.providerConfig() + `
resource "commercetools_connect_deployment" "sync" {
  key           = "missing"
  connector_key = "no-such-connector"
  region        = "europe-west1.gcp"
  application {
    name = "service"
  }
}
`,
				ExpectError: regexp.MustCompile(`connector not found`),
			},
			{
				Config: fake.providerConfig() + config("admin-key-1", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "deployment-evolve-sync"),
					resource.TestCheckResourceAttr(name, "status", "Deployed"),
					resource.TestCheckResourceAttr(name, "type", "sandbox"),
					resource.TestCheckResourceAttr(name, "version", "1"),
					resource.TestCheckResourceAttr(name, "application.0.url",
						"https://service-evolve-sync.europe-west1.gcp.commercetools.app"),
					resource.TestCheckNoResourceAttr(name, "application.1.schedule"),
					func(_ *terraform.State) error {
						_, secrets := fake.deployment("evolve-sync")
						assert.Equal(t, "admin-key-1", secrets["service/ALGOLIA_ADMIN_KEY"])
						return nil
					},
				),
			},
			{
				// A rotated secret and a new schedule redeploy in place.
				Config: fake.providerConfig() + config("admin-key-2", `schedule = "*/5 * * * *"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "deployment-evolve-sync"),
					resource.TestCheckResourceAttr(name, "version", "2"),
					resource.TestCheckResourceAttr(name, "status", "Deployed"),
					resource.TestCheckResourceAttr(name, "application.1.schedule", "*/5 * * * *"),
					func(_ *terraform.State) error {
						d, secrets := fake.deployment("evolve-sync")
						assert.Equal(t, "admin-key-2", secrets["service/ALGOLIA_ADMIN_KEY"])
						assert.Equal(t, "*/5 * * * *", *d.Applications[1].Schedule)
						return nil
					},
				),
			},
			{
				// A raised connector version redeploys in place.
				Config: fake.providerConfig() + configVersion("admin-key-2", `schedule = "*/5 * * * *"`, "connector_version = 4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "deployment-evolve-sync"),
					resource.TestCheckResourceAttr(name, "version", "3"),
					resource.TestCheckResourceAttr(name, "connector_version", "4"),
				),
			},
			{
				// A redeploy cannot move it to any other version.
				Config:      fake.providerConfig() + configVersion("admin-key-2", `schedule = "*/5 * * * *"`, "connector_version = 9"),
				ExpectError: regexp.MustCompile(`not to version 9`),
			},
			{
				Config: fake.providerConfig() + configVersion("admin-key-2", `schedule = "*/5 * * * *"`, "connector_version = 4"),
				Check:  resource.TestCheckResourceAttr(name, "connector_version", "4"),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     "evolve-sync",
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"application.0.secured_configuration",
					"application.0.secured_configuration.%",
					"application.0.secured_configuration.ALGOLIA_ADMIN_KEY",
				},
			},
		},
	})
}

func TestConnectDeploymentReleasedByCLI(t *testing.T) {
	defer connect_deployment.SetPollInterval(time.Millisecond)()
	fake := newFakeConnect(t)
	name := "commercetools_connect_deployment.sync"
	ignored := strings.Replace(configVersion("admin-key", "", "connector_version = 3"),
		"  application {", "  lifecycle {\n    ignore_changes = [connector_version]\n  }\n\n  application {", 1)

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.FrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fake.providerConfig() + ignored,
				Check:  resource.TestCheckResourceAttr(name, "connector_version", "3"),
			},
			{
				// The CLI redeployed it with --updateConnector; Terraform
				// keeps that version instead of moving it back.
				PreConfig: func() {
					fake.mu.Lock()
					defer fake.mu.Unlock()
					fake.deployments["evolve-sync"].Connector.Version = latestConnectorVersion
					fake.deployments["evolve-sync"].Version++
				},
				Config: fake.providerConfig() + ignored,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "connector_version", "4"),
					resource.TestCheckResourceAttr(name, "version", "2"),
				),
			},
			{
				Config:   fake.providerConfig() + ignored,
				PlanOnly: true,
			},
		},
	})
}

func TestConnectDeploymentFailure(t *testing.T) {
	defer connect_deployment.SetPollInterval(time.Millisecond)()
	fake := newFakeConnect(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.FrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				PreConfig:   func() { fake.failNext = true },
				Config:      fake.providerConfig() + config("admin-key", ""),
				ExpectError: regexp.MustCompile(`the deployment failed`),
			},
			{
				// The failed deployment was tainted, so this replaces it.
				Config: fake.providerConfig() + config("admin-key", ""),
				Check:  resource.TestCheckResourceAttr("commercetools_connect_deployment.sync", "status", "Deployed"),
			},
		},
	})
}
