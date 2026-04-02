# AGENTS.md — terraform-provider-commercetools

## Project Overview

Terraform provider for commercetools, written in Go 1.24. Uses a **dual-provider mux architecture**:
- `commercetools/` — Older resources using `terraform-plugin-sdk/v2`
- `internal/` — Newer resources using `terraform-plugin-framework`
- `main.go` muxes both via `tf5muxserver`

New resources must use the `terraform-plugin-framework` pattern in `internal/resources/`.

## Build & Run Commands

This project uses **Taskfile** (go-task), not Make.

| Command | Description |
|---|---|
| `task test` | Run unit tests (`go test ./...`) |
| `task testacc` | Start mock server, run all tests with `TF_ACC=true`, stop mock |
| `task format` | Run `go fmt ./...` and `terraform fmt` on examples |
| `task build-local` | Build and install to `~/.terraform.d/plugins/` (version 99.0.0) |
| `task docs` | Generate docs via `tfplugindocs` (`go generate`) |
| `task coverage` | Run tests with coverage, print function summary |

### Running a Single Test

```sh
# Single test by name (unit or acceptance):
go test -v -run TestAccState_createAndUpdateWithID ./internal/resources/state/

# Single test with acceptance mode (requires mock server running):
TF_ACC=true go test -v -run TestAccState_createAndUpdateWithID ./internal/resources/state/

# Start mock server first if running acceptance tests:
docker-compose up -d
TF_ACC=true go test -run TestAccChannel_AllFields ./commercetools/
docker-compose down
```

### Linting

```sh
golangci-lint run
```

Config: `.golangci.yml` — uses disable-all with explicit enables. Key linters: `goimports`, `govet`, `errcheck`, `staticcheck`, `unused`, `cyclop`, `forcetypeassert`.

## Project Structure

```
main.go                         # Entrypoint, provider mux
commercetools/                  # SDK-based resources (legacy)
  resource_<name>.go            # Resource implementation
  resource_<name>_test.go       # Tests
internal/
  provider/                     # Framework provider registration
  acctest/                      # Shared acceptance test helpers
  customtypes/                  # Custom TF types (LocalizedString)
  customvalidator/              # Custom TF validators
  models/                       # Shared data models
  sharedtypes/                  # Shared schema types (address, custom fields)
  utils/                        # Utilities (errors, refs, HCL templates, mutex)
  resources/<name>/             # Framework-based resources
    resource.go                 # CRUD + schema
    model.go                    # Model struct, NewXFromNative(), draft(), updateActions()
    resource_test.go            # Acceptance tests (external _test package)
    model_test.go               # Unit tests (same package, optional)
    upgrade_v<N>.go             # State migration (optional)
  datasource/<name>/            # Framework-based data sources
```

## Code Style

### Imports

Three groups separated by blank lines, enforced by `goimports`:
```go
import (
    "context"
    "time"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/labd/commercetools-go-sdk/platform"

    "github.com/labd/terraform-provider-commercetools/internal/utils"
)
```
1. Standard library
2. Third-party packages
3. Internal packages (full module path)

### Naming Conventions

- **Resource files (SDK):** `resource_<name>.go` — flat in `commercetools/`
- **Resource packages (Framework):** `internal/resources/<snake_case_name>/` with `resource.go`, `model.go`
- **Resource structs:** unexported camelCase — `stateResource`, `subscriptionResource`
- **Model structs:** exported PascalCase — `State`, `Subscription`, `AssociateRole`
- **Constructors:** `NewResource()` returns `resource.Resource`
- **Model constructors:** `NewStateFromNative(n *platform.State) State`
- **Model methods:** `draft()`, `updateActions(plan)`, `matchDefaults(state)`, `setDefaults()`
- **SDK CRUD:** `resourceChannelCreate`, `resourceChannelRead`, etc. (unexported)
- **Framework CRUD:** `Create()`, `Read()`, `Update()`, `Delete()` methods on resource struct
- **Test configs:** `testAccStateConfig(...)` — unexported, prefixed with `testAcc`
- **Destroy checks:** `testAccCheckStateDestroy`

### Resource Interface Checks

Every framework resource has compile-time interface satisfaction:
```go
var (
    _ resource.Resource                = &stateResource{}
    _ resource.ResourceWithConfigure   = &stateResource{}
    _ resource.ResourceWithImportState = &stateResource{}
)
```

### Error Handling

**Framework resources** — use diagnostics:
```go
diags := req.Plan.Get(ctx, &plan)
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() {
    return
}
```

**API errors** — use `AddError` with summary "Error <verb>ing <resource>":
```go
resp.Diagnostics.AddError("Error creating state", err.Error())
```

**404 handling in Read** — remove from state:
```go
if utils.IsResourceNotFoundError(err) {
    resp.State.RemoveResource(ctx)
    return
}
```

**All API calls** use retry with `utils.ProcessRemoteError()`:
```go
err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
    res, err = r.client.States().Post(draft).Execute(ctx)
    return utils.ProcessRemoteError(err)
})
```
- Create: `20*time.Second` timeout
- Update/Delete: `5*time.Second` timeout

### Schema Patterns

Every resource has `id` (Computed String) and `version` (Computed Int64). Use `tfsdk:"field_name"` struct tags (snake_case). Use `types.String`, `types.Int64`, `types.Bool` from the framework. For localized strings use `customtypes.LocalizedStringValue`. Slices use Go slices of type wrappers: `[]types.String`.

### Update Actions Pattern

Framework resources compare old state vs new plan directly using `reflect.DeepEqual` for complex types and `.Equal()` for simple types:
```go
func (s State) updateActions(plan State) platform.StateUpdate {
    result := platform.StateUpdate{Version: int(s.Version.ValueInt64())}
    if !reflect.DeepEqual(s.Name, plan.Name) {
        result.Actions = append(result.Actions, platform.StateSetNameAction{...})
    }
    return result
}
```

### Testing

- **Acceptance tests**: external test package (`_test` suffix), `resource.Test()` with `ProtoV5ProviderFactories` (framework) or `ProviderFactories` (SDK)
- **Unit tests**: same package for unexported method access; table-driven with `t.Run()`
- Use `github.com/stretchr/testify/assert` for assertions
- Use `utils.HCLTemplate(tmpl, map[string]any{...})` for test HCL configs

### Utility Functions

- `utils.OptionalString` / `utils.FromOptionalString` — `types.String` <-> `*string`
- `utils.Ref[T](T) *T` — generic pointer helper
- `utils.IsResourceNotFoundError(err)` — checks for 404
- `utils.ProcessRemoteError(err)` — classifies retry behavior

### Provider Data

Resources receive `*utils.ProviderData` via `Configure()`, providing `Client` (`*platform.ByProjectKeyRequestBuilder`) and `Mutex` (`*utils.MutexKV`).

## Environment Variables (Acceptance Tests)

`TF_ACC=true`, `CTP_CLIENT_ID`, `CTP_CLIENT_SECRET`, `CTP_PROJECT_KEY`, `CTP_SCOPES`, `CTP_API_URL`, `CTP_AUTH_URL`. For the mock server: API/Auth URL = `http://localhost:8989`.
