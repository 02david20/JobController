# Go Operator + Helm Distribution — Setup Template

> **Purpose**: A step-by-step reference for bootstrapping a Kubernetes operator with
> **Kubebuilder (Go)** and packaging it for **Helm** distribution.
> Derived from the `JobController` project structure.

---

## Table of Contents
0. [Quick Start](#0-quick-start)
1. [Prerequisites](#1-prerequisites)
2. [Initialise the Project](#2-initialise-the-project)
3. [Create API & Controller](#3-create-api--controller)
4. [Versioning Strategy](#4-versioning-strategy)
5. [Add Helm Chart](#5-add-helm-chart)
6. [Makefile Targets Reference](#6-makefile-targets-reference)
7. [Local Testing with Kind](#7-local-testing-with-kind)
8. [CI/CD — GitHub Actions Pipeline](#8-cicd--github-actions-pipeline)
9. [Distribution via OCI Registry](#9-distribution-via-oci-registry)
10. [Project Structure Reference](#10-project-structure-reference)
11. [Quick-Reference Cheat Sheet](#11-quick-reference-cheat-sheet)

---

## 0. Quick Start

> Get from zero to a running operator with Helm in under 10 minutes.
> Replace all `<placeholder>` values with your own before running.

```bash
# ── Step 1: Bootstrap the project ────────────────────────────────────────────
mkdir my-operator && cd my-operator
go mod init github.com/<your-org>/my-operator

kubebuilder init \
  --domain <your-domain.io> \
  --repo github.com/<your-org>/my-operator

# ── Step 2: Scaffold API + controller ────────────────────────────────────────
kubebuilder create api \
  --group <group> \
  --version v1alpha1 \
  --kind <Kind> \
  --resource true \
  --controller true

# Regenerate CRDs & DeepCopy after editing *_types.go
make manifests generate

# ── Step 3: Add Helm chart ───────────────────────────────────────────────────
kubebuilder edit --plugins=helm.kubebuilder.io/v2-alpha
# Chart is now at dist/chart/

# ── Step 4: Run tests ────────────────────────────────────────────────────────
make test                 # Unit tests (no cluster needed)
make helm-chart           # Sync chart from Kustomize
make helm-test            # Helm unit tests

# ── Step 5: Build & load image into a local Kind cluster ─────────────────────
kind create cluster --name my-operator-dev
export IMG=my-operator:dev
make docker-build IMG=$IMG
kind load docker-image $IMG --name my-operator-dev

# ── Step 6: Deploy via Helm ───────────────────────────────────────────────────
make helm-deploy IMG=$IMG
kubectl get all -n my-operator-system

# ── Step 7: Verify controller logs ───────────────────────────────────────────
kubectl logs -n my-operator-system \
  deployment/my-operator-controller-manager -c manager -f
```

> See the sections below for detailed explanations of each step.

---

## 1. Prerequisites

| Tool               | Minimum Version | Install (macOS)                     |
| --------------------| -----------------| -------------------------------------|
| Go                 | 1.22+           | `brew install go`                   |
| Docker             | latest          | `brew install --cask docker`        |
| Kubebuilder CLI    | v4.x            | `brew install kubebuilder`          |
| Helm               | v3              | `brew install helm`                 |
| Kind               | latest          | `brew install kind`                 |
| kubectl            | v1.28+          | `brew install kubectl`              |
| Container Registry | —               | Docker Hub / GHCR / Azure CR / etc. |

> **Tip**: Verify Kubebuilder is on your `$PATH`:
> ```bash
> kubebuilder version
> # Should print: Version: v4.x.x ...
> ```

---

## 2. Initialise the Project

```bash
# 1. Create the project directory
mkdir my-operator && cd my-operator

# 2. Initialise the Go module (replace with your actual repo URL)
go mod init github.com/<your-org>/my-operator

# 3. Scaffold the Kubebuilder project
kubebuilder init \
  --domain <your-domain.io> \
  --repo github.com/<your-org>/my-operator
```

This scaffolds the standard single-group layout:

```
cmd/main.go                 # Manager entry point
api/                        # CRD type definitions (empty at this stage)
internal/controller/        # Controller logic (empty at this stage)
config/                     # Kustomize manifests (CRDs, RBAC, manager)
Makefile                    # Build, test, deploy commands
Dockerfile                  # Multi-stage build for the manager image
go.mod / go.sum             # Go module files
PROJECT                     # Kubebuilder metadata (DO NOT EDIT)
```

### Optional: Multi-Group Layout

If you need multiple API groups (e.g., `batch`, `apps`), convert before scaffolding any APIs:

```bash
kubebuilder edit --multigroup=true
```

Then APIs live at `api/<group>/<version>/`, controllers at `internal/controller/<group>/`.

---

## 3. Create API & Controller

```bash
kubebuilder create api \
  --group <group> \
  --version v1alpha1 \
  --kind <Kind> \
  --resource true \
  --controller true
```

After editing `api/v1alpha1/<kind>_types.go`, always regenerate:

```bash
make manifests   # Regenerate CRDs & RBAC from +kubebuilder markers
make generate    # Regenerate DeepCopy methods
```

### Key Markers for `*_types.go`

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.conditions[?(@.type=='Ready')].status"

// Field-level markers:
// +kubebuilder:validation:Required
// +kubebuilder:validation:Minimum=1
// +kubebuilder:default="value"
```

> ⚠️ **Never delete `// +kubebuilder:scaffold:*` comments** — the CLI injects code at these markers.

### After Editing Controller Logic

```bash
make lint-fix    # Auto-fix code style
make test        # Run unit tests (uses envtest)
```

---

## 4. Versioning Strategy

This project uses a **`VERSION` file + git tag counter** approach.

**`VERSION` file** contains the `major.minor` base (e.g., `0.1`).

In `Makefile`, the full version is computed automatically:

```makefile
VERSION_BASE := $(shell cat VERSION 2>/dev/null || echo "0.1")
PATCH        := $(shell git tag -l "$(VERSION_BASE).*" 2>/dev/null | wc -l | tr -d ' ')
VERSION      := $(VERSION_BASE).$(PATCH)
```

The resulting version is `<major>.<minor>.<patch>` where `PATCH` auto-increments with each tag.

**To bump the patch version:**
```bash
make release-tag       # Tags the current commit as $(VERSION)
git push origin <TAG>  # Push the tag to trigger CI
```

**To bump major/minor:** Edit the `VERSION` file manually (e.g., `0.1` → `0.2`).

**In CI**, the version is derived from the `VERSION` file and `github.run_number`:
```bash
BASE_VERSION=$(cat VERSION | tr -d '[:space:]')
MAJOR_MINOR=$(echo "$BASE_VERSION" | cut -d'.' -f1,2)
APP_VERSION="${MAJOR_MINOR}.${GITHUB_RUN_NUMBER}"
```

---

## 5. Helm Distribution

Helm is the **primary distribution mechanism** for this project. The Kubebuilder Helm plugin
keeps your `dist/chart/` automatically in sync with your Kustomize config — you never have to
hand-edit templates.

### 5.1 First-Time Setup

```bash
# Register the Helm plugin and generate dist/chart/ for the first time
kubebuilder edit --plugins=helm.kubebuilder.io/v2-alpha
```

This registers `helm.kubebuilder.io/v2-alpha` in the `PROJECT` file and writes the initial
chart. After this, use `make helm-chart` (or `make generate-dist`) for all future syncs.

The generated chart lives at `dist/chart/`:

```
dist/chart/
├── Chart.yaml              # Chart name, version, appVersion, description
├── values.yaml             # User-configurable values
├── values.schema.json      # JSON Schema for values (generated — never edit manually)
├── .helmignore             # Files to exclude from helm package
├── templates/              # Kubernetes manifests (generated from config/)
└── tests/                  # Helm unit tests (helm-unittest)
    ├── deployment_snapshot_test.yaml
    ├── metrics_snapshot_test.yaml
    ├── rbac_snapshot_test.yaml
    └── __snapshot__/       # Snapshot baselines (commit to git)
```

---

### 5.2 `make generate-dist` — Full Distribution Refresh

`generate-dist` is the **all-in-one command** for keeping the distribution artifacts in sync.
Run it whenever you:
- Add or change a CRD field / RBAC rule
- Modify the manager Deployment config
- Want to cut a new release

```bash
make generate-dist
```

**What it does, in order:**

| Step | Command | Effect |
|------|---------|--------|
| 1 | `make build-installer` | Generates `dist/install.yaml` from Kustomize |
| 2 | `kubebuilder edit --plugins=helm.kubebuilder.io/v2-alpha` | Syncs `dist/chart/templates/` from `dist/install.yaml` |
| 3 | `make helm-schema` | Regenerates `dist/chart/values.schema.json` from `values.yaml` |
| 4 | `helm lint dist/chart` | Validates the chart before committing |
| 5 | `make helm-update-snapshots` | Refreshes snapshot baselines for unit tests |

> **Rule of thumb**: run `make generate-dist` before every commit that touches
> `api/`, `config/`, or `internal/controller/`.

---

### 5.3 `make helm-chart` — Lightweight Chart Sync

Use this for a faster sync that skips the snapshot update step:

```bash
make helm-chart
```

**What it does:**

| Step | Command | Effect |
|------|---------|--------|
| 1 | `kubebuilder edit --plugins=helm.kubebuilder.io/v2-alpha` | Syncs `dist/chart/templates/` |
| 2 | `make helm-schema` | Regenerates `values.schema.json` |

Typical use: after changing controller RBAC or config — quick sync before running `make helm-test`.

---

### 5.4 `values.yaml` Customisation

After the first `generate-dist`, customise `dist/chart/values.yaml` to expose operator-specific
knobs to end users:

```yaml
# dist/chart/values.yaml (example)
manager:
  image:
    repository: your-registry/my-operator   # set by CI during release
    tag: "0.1.0"
  replicas: 1
  resources:
    limits:
      cpu: 500m
      memory: 128Mi
    requests:
      cpu: 10m
      memory: 64Mi
```

After editing `values.yaml`, regenerate the schema:

```bash
make helm-schema   # Writes dist/chart/values.schema.json
```

---

### 5.5 Helm Unit Tests

The chart ships with snapshot tests (via [helm-unittest](https://github.com/helm-unittest/helm-unittest)).
These catch accidental changes to generated manifests.

```bash
make helm-test              # Run snapshot tests
make helm-update-snapshots  # Regenerate baselines after intentional changes
```

Snapshot files live in `dist/chart/tests/__snapshot__/` — **commit them to git**.

---

### 5.6 Package the Chart

To produce a distributable `.tgz`:

```bash
make helm-package
# Outputs: dist/my-operator-<version>.tgz
```

This uses the auto-computed `$(VERSION)` and `$(CHART_VERSION)` from the `Makefile`.
To override:

```bash
CHART_VERSION=1.2.3 make helm-package
```

> Commit the `__snapshot__/` directory — CI uses it to detect regressions.

---

## 6. Makefile Targets Reference

### Development

| Target | Description |
|--------|-------------|
| `make manifests` | Regenerate CRDs & RBAC from `+kubebuilder` markers |
| `make generate` | Regenerate `DeepCopy` methods |
| `make fmt` | `go fmt ./...` |
| `make vet` | `go vet ./...` |
| `make lint` | Run `golangci-lint` |
| `make lint-fix` | Run `golangci-lint --fix` |
| `make lint-config` | Verify linter config |

### Build & Image

| Target | Description |
|--------|-------------|
| `make build` | Build manager binary to `bin/manager` |
| `make docker-build IMG=<img>` | Build Docker image |
| `make docker-push IMG=<img>` | Push Docker image |
| `make docker-buildx IMG=<img>` | Multi-platform build & push |
| `make build-installer IMG=<img>` | Generate `dist/install.yaml` (Kustomize bundle) |

### Testing

| Target | Description |
|--------|-------------|
| `make test` | Unit tests with envtest (no cluster needed) |
| `make test-e2e` | E2E tests in an isolated Kind cluster |
| `make test-helm-e2e` | Helm E2E tests in an isolated Kind cluster |
| `make helm-test` | Helm unit tests with `helm unittest` |
| `make helm-update-snapshots` | Regenerate Helm snapshot baselines |

### Helm — Distribution

| Target | Description |
|--------|-------------|
| `make generate-dist` | **Full refresh**: build-installer → helm-chart → helm-schema → lint → update-snapshots |
| `make helm-chart` | Lightweight sync: re-run Helm plugin + regenerate schema |
| `make helm-schema` | Generate `dist/chart/values.schema.json` from `values.yaml` |
| `make helm-package` | Package `dist/chart/` as a versioned `.tgz` in `dist/` |
| `make helm-test` | Run Helm unit/snapshot tests (`helm unittest`) |
| `make helm-update-snapshots` | Regenerate snapshot baselines after intentional changes |

### Helm — Local Development

| Target | Description |
|--------|-------------|
| `make helm-deploy IMG=<img>` | `helm upgrade --install` on the current cluster |
| `make helm-uninstall` | Uninstall the Helm release |
| `make helm-status` | Show release status |
| `make helm-history` | Show release history |
| `make helm-rollback` | Rollback to previous release |

### Versioning

| Target | Description |
|--------|-------------|
| `make release-tag` | Git tag current commit as `$(VERSION)` |

---

## 7. Local Testing with Kind

### Unit Tests (No Cluster Required)

```bash
make test
```

Uses `controller-runtime/tools/setup-envtest` to spin up a real Kubernetes API server + etcd in-process.

### E2E Tests (Isolated Kind Cluster)

```bash
make test-e2e
# Internally: creates kind cluster "my-operator-test-e2e", runs tests, then deletes cluster
```

To use a custom cluster name:
```bash
KIND_CLUSTER=my-cluster make test-e2e
```

### Helm E2E Tests (Isolated Kind Cluster)

```bash
make test-helm-e2e
# Internally: creates kind cluster "my-operator-test-helm-e2e", installs via Helm, runs tests, deletes cluster
```

### Local Helm Development Workflow (Inner Loop)

A typical development cycle using a local Kind cluster:

```bash
# ── 1. Create a dedicated dev cluster ────────────────────────────────────────
kind create cluster --name my-operator-dev

# ── 2. Sync chart after any config / API change ───────────────────────────────
make generate-dist     # Full refresh (CRDs + chart + schema + lint + snapshots)
# or for a quicker sync:
make helm-chart        # Chart + schema only

# ── 3. Build & load the manager image (no registry needed) ───────────────────
export IMG=my-operator:dev
make docker-build IMG=$IMG
kind load docker-image $IMG --name my-operator-dev

# ── 4. Install / upgrade via Helm ─────────────────────────────────────────────
make helm-deploy IMG=$IMG
# Runs: helm upgrade --install my-operator dist/chart \
#         --namespace my-operator-system --create-namespace \
#         --set manager.image.repository=my-operator \
#         --set manager.image.tag=dev \
#         --wait --timeout 5m

# ── 5. Verify the deployment ──────────────────────────────────────────────────
kubectl get all -n my-operator-system
make helm-status

# ── 6. Tail controller logs ───────────────────────────────────────────────────
kubectl logs -n my-operator-system \
  deployment/my-operator-controller-manager -c manager -f

# ── 7. Iterate: code change → rebuild → redeploy ─────────────────────────────
make docker-build IMG=$IMG
kind load docker-image $IMG --name my-operator-dev
make helm-deploy IMG=$IMG          # Helm upgrade re-uses the same release name

# ── 8. Rollback if needed ─────────────────────────────────────────────────────
make helm-history                  # See revision list
make helm-rollback                 # Roll back to previous revision

# ── 9. Teardown ───────────────────────────────────────────────────────────────
make helm-uninstall
kind delete cluster --name my-operator-dev
```

**Override default Helm variables** without editing the Makefile:

```bash
HELM_NAMESPACE=staging \
HELM_RELEASE=my-operator-staging \
HELM_EXTRA_ARGS="--set manager.replicas=2" \
make helm-deploy IMG=$IMG
```

---

## 8. CI/CD — GitHub Actions Pipeline

This project uses a **five-job pipeline** that gates the release on all test stages passing:

```
lint ──────────┐
test ──────────┤
test-chart ────┤──► release  (main branch / workflow_dispatch only)
test-e2e ──────┤
test-helm-e2e ─┘
```

### Required Repository Variables & Secrets

| Name | Type | Description |
|------|------|-------------|
| `REGISTRY` | Variable | Registry host or path (e.g., `docker.io`, `myrepo.azurecr.io`) |
| `REGISTRY_USERNAME` | Secret | Registry login username |
| `REGISTRY_PASSWORD` | Secret | Registry login password or token |

### Full Pipeline Workflow

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
  workflow_dispatch:

env:
  REGISTRY: ${{ vars.REGISTRY || 'docker.io' }}
  REGISTRY_USERNAME: ${{ secrets.REGISTRY_USERNAME }}
  REGISTRY_PASSWORD: ${{ secrets.REGISTRY_PASSWORD }}
  IMAGE_NAME: my-operator-manager

jobs:
  # ── 1. Lint ─────────────────────────────────────────────────────────
  lint:
    name: Lint Code
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22', cache: true }
      - run: make lint-config
      - run: make lint

  # ── 2. Unit Tests ───────────────────────────────────────────────────
  test:
    name: Run Unit Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22', cache: true }
      - run: go mod tidy && make test

  # ── 3. Helm Chart Lint & Unit Tests ─────────────────────────────────
  test-chart:
    name: Lint & Test Helm Chart
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22', cache: true }
      - uses: azure/setup-helm@v4
        with: { version: 'v3.19.0' }
      - run: helm plugin install https://github.com/helm-unittest/helm-unittest
      - run: make helm-chart
      - run: helm lint ./dist/chart
      - run: make helm-test

  # ── 4. Operator E2E Tests ────────────────────────────────────────────
  test-e2e:
    name: Run E2E Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22', cache: true }
      - name: Install Kind
        run: |
          curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-$(go env GOARCH)
          chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind
      - run: go mod tidy && make test-e2e

  # ── 5. Helm E2E Tests ────────────────────────────────────────────────
  test-helm-e2e:
    name: Run Helm E2E Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22', cache: true }
      - uses: azure/setup-helm@v4
        with: { version: 'v3.19.0' }
      - name: Install Kind
        run: |
          curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-$(go env GOARCH)
          chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind
      - run: make helm-chart
      - run: go mod tidy && make test-helm-e2e

  # ── 6. Release (main branch only) ────────────────────────────────────
  release:
    name: Build and Push Release
    needs: [lint, test, test-chart, test-e2e, test-helm-e2e]
    if: (github.event_name == 'push' && github.ref == 'refs/heads/main') || github.event_name == 'workflow_dispatch'
    runs-on: ubuntu-latest
    environment: Build
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22', cache: true }
      - uses: azure/setup-helm@v4
        with: { version: 'v3.19.0' }
      - uses: docker/setup-buildx-action@v3

      - name: Parse Registry Metadata & Compute Version
        run: |
          RAW_REG="${{ env.REGISTRY }}"
          RAW_USER="${{ env.REGISTRY_USERNAME }}"
          IMG_NAME="${{ env.IMAGE_NAME }}"

          FIRST_PART=$(echo "$RAW_REG" | cut -d'/' -f1)
          if [[ "$FIRST_PART" == *.* ]]; then
            REGISTRY_HOST="$FIRST_PART"
            REG_PATH=$(echo "$RAW_REG" | cut -d'/' -f2-)
            [ "$REG_PATH" = "$FIRST_PART" ] && REG_PATH=""
          else
            REGISTRY_HOST="docker.io"
            REG_PATH="$RAW_REG"
          fi

          if [ -z "$REG_PATH" ]; then
            DOCKER_IMAGE_PRIMARY="${REGISTRY_HOST}/${RAW_USER}/${IMG_NAME}"
            HELM_REPO_URL="oci://${REGISTRY_HOST}/${RAW_USER}"
          elif [[ "$REG_PATH" == */* ]]; then
            DOCKER_IMAGE_PRIMARY="${REGISTRY_HOST}/${REG_PATH}"
            HELM_REPO_URL="oci://${REGISTRY_HOST}/$(dirname "$REG_PATH")"
          else
            DOCKER_IMAGE_PRIMARY="${REGISTRY_HOST}/${REG_PATH}/${IMG_NAME}"
            HELM_REPO_URL="oci://${REGISTRY_HOST}/${REG_PATH}"
          fi

          BASE_VERSION=$(cat VERSION | tr -d '[:space:]')
          MAJOR_MINOR=$(echo "$BASE_VERSION" | cut -d'.' -f1,2)
          APP_VERSION="${MAJOR_MINOR}.${{ github.run_number }}"

          echo "REGISTRY_HOST=$REGISTRY_HOST"               >> $GITHUB_ENV
          echo "DOCKER_IMAGE_PRIMARY=$DOCKER_IMAGE_PRIMARY" >> $GITHUB_ENV
          echo "HELM_REPO_URL=$HELM_REPO_URL"               >> $GITHUB_ENV
          echo "APP_VERSION=$APP_VERSION"                   >> $GITHUB_ENV

      - name: Log in to Docker Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY_HOST }}
          username: ${{ env.REGISTRY_USERNAME }}
          password: ${{ env.REGISTRY_PASSWORD }}

      - name: Log in to Helm OCI Registry
        run: |
          echo "${{ env.REGISTRY_PASSWORD }}" | \
          helm registry login ${{ env.REGISTRY_HOST }} \
            --username ${{ env.REGISTRY_USERNAME }} --password-stdin

      - name: Sync Helm Chart
        run: make helm-chart

      - name: Build & Push Manager Image
        run: |
          make docker-build IMG=${{ env.DOCKER_IMAGE_PRIMARY }}:${{ env.APP_VERSION }}
          docker tag ${{ env.DOCKER_IMAGE_PRIMARY }}:${{ env.APP_VERSION }} \
                     ${{ env.DOCKER_IMAGE_PRIMARY }}:latest
          docker tag ${{ env.DOCKER_IMAGE_PRIMARY }}:${{ env.APP_VERSION }} \
                     ${{ env.DOCKER_IMAGE_PRIMARY }}:${{ github.sha }}
          make docker-push IMG=${{ env.DOCKER_IMAGE_PRIMARY }}:${{ env.APP_VERSION }}
          make docker-push IMG=${{ env.DOCKER_IMAGE_PRIMARY }}:latest
          make docker-push IMG=${{ env.DOCKER_IMAGE_PRIMARY }}:${{ github.sha }}

      - name: Package & Push Helm Chart to OCI Registry
        run: |
          # Inject the built image into chart values
          sed -i "s|repository: controller|repository: ${DOCKER_IMAGE_PRIMARY}|g" \
            dist/chart/values.yaml

          helm package dist/chart \
            --version ${{ env.APP_VERSION }} \
            --app-version ${{ env.APP_VERSION }}

          helm push my-operator-${{ env.APP_VERSION }}.tgz ${{ env.HELM_REPO_URL }}
```

---

## 9. Distribution via OCI Registry

This project publishes both the **container image** and **Helm chart** as OCI artifacts.

### Install from OCI Registry (End Users)

```bash
# Authenticate
helm registry login <registry-host> --username <user> --password <token>

# Install directly from OCI
helm install my-operator \
  oci://<registry-host>/<registry-path>/my-operator \
  --version 0.1.42 \
  --namespace my-operator-system \
  --create-namespace
```

### Alternative: Kustomize YAML Bundle

For users without Helm, generate a single self-contained manifest:

```bash
make build-installer IMG=<registry>/<image>:<tag>
# Outputs: dist/install.yaml

# End users install with:
kubectl apply -f dist/install.yaml
```

---

## 10. Project Structure Reference

```
my-operator/
├── api/
│   └── v1alpha1/
│       ├── <kind>_types.go          # CRD schema (+kubebuilder markers)
│       └── zz_generated.deepcopy.go # Auto-generated (make generate)
├── cmd/
│   └── main.go                      # Manager entry point
├── config/
│   ├── crd/bases/                   # Auto-generated CRDs (make manifests)
│   ├── rbac/                        # Auto-generated RBAC (make manifests)
│   ├── manager/                     # Manager Deployment kustomize patch
│   └── default/                     # Default kustomize overlay
├── dist/
│   ├── install.yaml                 # Kustomize bundle (make build-installer)
│   └── chart/                       # Helm chart
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── values.schema.json       # Generated (make helm-schema)
│       ├── .helmignore
│       ├── templates/               # Rendered manifests
│       └── tests/                   # Helm unit tests
│           └── __snapshot__/        # Baselines — commit these!
├── hack/
│   └── boilerplate.go.txt           # License header for generated files
├── internal/
│   └── controller/
│       ├── <kind>_controller.go     # Reconciliation logic
│       └── suite_test.go            # envtest bootstrap
├── test/
│   ├── e2e/                         # Operator E2E tests (Kind)
│   └── helm_e2e/                    # Helm E2E tests (Kind)
├── docs/
│   └── SETUP_TEMPLATE.md            # This file
├── .github/
│   └── workflows/
│       └── release.yml              # CI/CD pipeline
├── .golangci.yml                    # Linter configuration
├── .custom-gcl.yml                  # Custom golangci-lint plugins (optional)
├── Dockerfile                       # Multi-stage build
├── Makefile                         # All build/test/deploy commands
├── PROJECT                          # Kubebuilder metadata (DO NOT EDIT)
└── VERSION                          # Base version file (e.g., "0.1")
```

---

## 11. Quick-Reference Cheat Sheet

| Goal | Command |
|------|---------|
| Init project | `kubebuilder init --domain example.io --repo github.com/org/my-operator` |
| Scaffold API | `kubebuilder create api --group mygroup --version v1alpha1 --kind MyKind --resource --controller` |
| Scaffold webhook | `kubebuilder create webhook --group mygroup --version v1alpha1 --kind MyKind --defaulting --programmatic-validation` |
| Regenerate CRDs/RBAC | `make manifests` |
| Regenerate DeepCopy | `make generate` |
| Fix code style | `make lint-fix` |
| Run unit tests | `make test` |
| Run E2E tests | `make test-e2e` |
| Run Helm E2E tests | `make test-helm-e2e` |
| Build image | `make docker-build IMG=<img>` |
| Push image | `make docker-push IMG=<img>` |
| Sync Helm chart | `make helm-chart` |
| Run Helm unit tests | `make helm-test` |
| Update Helm snapshots | `make helm-update-snapshots` |
| Local Helm deploy | `make helm-deploy IMG=<img>` |
| Package Helm chart | `make helm-package` |
| Bump version tag | `make release-tag && git push origin <TAG>` |

---

## References

- [Kubebuilder Book](https://book.kubebuilder.io)
- [controller-runtime FAQ](https://github.com/kubernetes-sigs/controller-runtime/blob/main/FAQ.md)
- [Helm Chart Best Practices](https://helm.sh/docs/chart_best_practices/)
- [helm-unittest](https://github.com/helm-unittest/helm-unittest)
- [helm-values-schema-json](https://github.com/losisin/helm-values-schema-json)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Logging Message Style Guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md#message-style-guidelines)
