# Helm Distribution & Deployment Guide

This guide details how to build, package, test, and deploy the **JobController** operator using our modern, official Kubebuilder Helm plugin pipeline.

---

## 1. Architecture: Kubebuilder Helm Plugin (`helm/v2-alpha`)

Instead of writing a manual monolithic Helm template or relying on custom python scripts, the project uses the official **Kubebuilder Helm Plugin** (`helm/v2-alpha`). 

This setup allows us to:
1. Maintain raw Kubernetes manifests under `config/` (CRDs, RBAC roles, manager Deployment) as the single source of truth.
2. Compile and compile Kustomize resources into **clean, split-template files** within our Helm chart directory (`dist/chart/`).
3. Leverage automatic Helm schema generation to validate values file changes.

### Helm Chart Directory Structure
The Helm chart source is located under `dist/chart/` and organized as follows:

```text
dist/chart/
├── Chart.yaml                  # Chart versioning and metadata
├── values.yaml                 # User-overridable configuration variables
├── values.schema.json          # Automatically generated JSON schema for values.yaml
├── templates/                  # Split template files compiled from Kustomize
│   ├── NOTES.txt               # Installation instructions
│   ├── _helpers.tpl            # Common template helpers and labels
│   ├── crd/                    # CRDs (e.g. JobRunner CRD)
│   ├── manager/                # Manager Deployment configuration
│   ├── metrics/                # Metrics Service configuration
│   ├── prometheus/             # Prometheus ServiceMonitor configuration
│   └── rbac/                   # Group-specific and cluster RBAC roles & bindings
└── tests/                      # Unit/snapshot tests
    ├── __snapshots__/          # Reference snapshot baselines
    ├── deployment_snapshot_test.yaml
    ├── metrics_snapshot_test.yaml
    └── rbac_snapshot_test.yaml
```

---

## 2. Values.yaml Key Parameters

The Helm chart exposes standard parameters in `dist/chart/values.yaml` to customize deployment:

| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `controllerManager.manager.image.repository` | string | `controller` | Docker image repository. Replaced dynamically during push. |
| `controllerManager.manager.image.tag` | string | `latest` | Docker image tag. Replaced dynamically during push. |
| `metrics.enable` | boolean | `false` | Whether to enable the metrics service port. If `true`, a Metrics Service and ServiceMonitor are generated. |
| `kubernetesClusterDomain` | string | `cluster.local` | Domain for service discovery resolution. |

---

## 3. Development & Management (Makefile Targets)

We provide Makefile targets to handle building, schema updates, local testing, and packaging:

### Syncing Kustomize to Helm
To compile Kustomize modifications into the split Helm templates, run:
```bash
make helm-chart
```
*Under the hood, this runs `kubebuilder edit --plugins=helm.kubebuilder.io/v2-alpha` to populate `dist/chart/templates/` and generates the JSON schema.*

### Rebuilding & Linting All Assets
To build the installer manifest, sync the Helm chart, regenerate the schema, and lint the resulting templates:
```bash
make generate-dist
```

### Running Unit/Snapshot Tests
Unit tests use the `helm-unittest` plugin to assert template properties and prevent regression against snapshot baselines:
```bash
make helm-test
```

### Updating Test Snapshot Baselines
If you change Kustomize manifests intentionally and need to update the unit test snapshots to match:
```bash
make helm-update-snapshots
```

### Running Helm E2E Integration Tests
To spin up an isolated local Kind cluster, deploy the operator via Helm, and run validation specs (asserting the controller manager and metrics service are running):
```bash
make test-helm-e2e
```

---

## 4. Automatic Versioning & CI/CD Pipeline

The GitHub Actions release pipeline ([`.github/workflows/release.yml`](file:///Users/david/Documents/Go/JobController/.github/workflows/release.yml)) automates tagging, building, and publishing:

1. **VERSION File**: The base version is specified in the [VERSION](file:///Users/david/Documents/Go/JobController/VERSION) file (e.g. `0.1.0`).
2. **Dynamic Version Calculation**: 
   - The CI pipeline parses the `major.minor` components from `VERSION`.
   - It appends the GitHub Actions run number as the patch version (e.g., `0.1.42` for run 42).
3. **Tag Syncing**:
   - The manager Docker image is built and pushed to the registry as `davidp0c/jobcontroller-manager:<version>`.
   - In the packaged Helm chart, `sed` dynamically replaces the repository and tags in `values.yaml` to target the built image.
   - The Helm chart is packaged as `jobcontroller-<version>.tgz` and pushed to the Docker Hub OCI registry.

---

## 5. CLI Installation Guide

To deploy the operator from the remote OCI registry:

### Step 1: Create the Namespace
```bash
kubectl create namespace jobcontroller-system
```

### Step 2: Configure Pull Credentials
If pulling the controller manager image from a private repository, create the `docker-registry` secret:
```bash
kubectl create secret docker-registry jobcontroller-pull-secret \
  --docker-server=https://index.docker.io/v1/ \
  --docker-username="davidp0c" \
  --docker-password="YOUR_DOCKER_HUB_ACCESS_TOKEN_OR_PASSWORD" \
  --docker-email="your-email@example.com" \
  --namespace jobcontroller-system
```

### Step 3: Authenticate your Helm Client
Log in to the Docker Hub registry using your credentials:
```bash
echo "YOUR_DOCKER_HUB_PASSWORD" | helm registry login registry-1.docker.io --username davidp0c --password-stdin
```

### Step 4: Install or Upgrade the Operator
```bash
helm upgrade --install jobcontroller oci://registry-1.docker.io/davidp0c/jobcontroller \
  --version <VERSION> \
  --namespace jobcontroller-system
```

---

## 6. GitOps Deployment via Argo CD

You can deploy the operator through Argo CD using either a Git repository reference or the Helm OCI registry path.

### Method A: Deploying from the Git Repository (Subdirectory)
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: job-controller
  namespace: argocd
spec:
  project: default
  source:
    repoURL: 'https://github.com/02david20/JobController'
    targetRevision: main
    path: dist/chart
    helm:
      valueFiles:
        - values.yaml
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: jobcontroller-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    createNamespace: true
```

### Method B: Deploying from your OCI Helm Registry
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: job-controller
  namespace: argocd
spec:
  project: default
  source:
    chart: jobcontroller
    repoURL: 'registry-1.docker.io/davidp0c'
    targetRevision: 0.1.42  # Chart version
    helm:
      parameters:
        - name: metrics.enable
          value: "true"
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: jobcontroller-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    createNamespace: true
```
