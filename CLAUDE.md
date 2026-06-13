# CLAUDE.md — Project Guidelines & Development Context

This document provides quick-reference commands and critical guidelines for working with the **JobController** operator project.

---

## 1. Useful Development Commands

### Code Generation & Linting
- **Regenerate CRDs and deepcopy code**: `make manifests generate`
- **Run linter**: `make lint`
- **Auto-fix lint issues**: `make lint-fix`

### Running Tests
- **Run standard unit tests**: `make test` (requires envtest)
- **Run Helm template unit tests**: `make helm-test` (requires `helm-unittest` plugin)
- **Update Helm unit test snapshot baselines**: `make helm-update-snapshots`
- **Run local standard E2E tests**: `make test-e2e` (creates and cleans up a Kind cluster)
- **Run local Helm integration E2E tests**: `make test-helm-e2e` (creates and cleans up a Kind cluster)

### Local Development Running
- **Run operator manager locally**: `make run` (uses your current kubeconfig context)

---

## 2. Helm Chart Pipeline (`dist/chart/`)

The Helm chart is located under [dist/chart/](file:///Users/david/Documents/Go/JobController/dist/chart/) and is dynamically compiled from Kustomize templates.

- **Do NOT hand-edit template files** inside `dist/chart/templates/`. They are regenerated and overwritten when running:
  ```bash
  make helm-chart
  ```
- **Files safe to edit manually**:
  - `dist/chart/Chart.yaml` (metadata & chart definition)
  - `dist/chart/values.yaml` (overridable values)
  - `dist/chart/tests/` (unit and snapshot test YAML files)
- **Values JSON Schema**: The `values.schema.json` is auto-generated. When modifying `values.yaml`, regenerate the schema via:
  ```bash
  make helm-schema
  ```

---

## 3. CI/CD & GitHub Actions Pipelines

Unlike the reference Azure DevOps pipeline, this project uses **GitHub Actions** located under [`.github/workflows/`](file:///Users/david/Documents/Go/JobController/.github/workflows/):
- **`release.yml`**: Triggers on push to `main` and pull requests. Runs lint, test, test-chart, test-e2e, and test-helm-e2e. Builds and pushes manager images and OCI Helm packages on merge to `main`.
- **`docs.yml`**: Builds and deploys documentation to GitHub Pages.

> [!WARNING]
> The Kubebuilder Helm plugin (`helm/v2-alpha`) unconditionally generates a root `.github/workflows/test-chart.yml` file when syncing templates. Our Makefile targets automatically delete this file (`rm -f .github/workflows/test-chart.yml`) to preserve our customized CI/CD workflow. Avoid manual staging of `test-chart.yml`.

---

## 4. Release Versioning

- The base version is specified in the [VERSION](file:///Users/david/Documents/Go/JobController/VERSION) file (e.g. `0.1.0`).
- During release workflows, the pipeline dynamically calculates the version tag as `major.minor.RUN_NUMBER` (using the GitHub actions run number) and syncs both the container image tag and the Helm chart tag.
- To tag a release manually on your local system:
  ```bash
  make release-tag
  ```
