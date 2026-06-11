# Helm Distribution & Deployment Guide

This guide details how to build, package, and deploy the JobController operator using our decoupled Helm chart pipeline.

## Decoupled Pipeline (cert-manager Pattern)
Instead of combining Helm's text-templating logic directly with Kustomize (which breaks YAML syntax validators and makes it difficult to maintain), we use a decoupled design:
1. **Source Code & Manifests**: We maintain raw manifests via Kubebuilder and Kustomize under the `config/` folder.
2. **Sync Script**: When building releases, we compile Kustomize outputs into single static files inside the Helm chart templates folder. This ensures Kustomize remains the single source of truth.

---

## Syncing Kustomize to Helm
Whenever you make changes to Go APIs, RBAC tags, or manager deployment configurations, sync the updates directly to your Helm chart using the following Makefile command:

```bash
make sync-helm
```

### What this does under the hood:
- Automatically rebuilds the latest Kustomize outputs.
- Creates/updates `deploy/charts/jobcontroller/templates/manifests.yaml` (which contains deployment, RBAC roles, and service account manifests).
- Creates/updates `deploy/charts/jobcontroller/crds/crds.yaml` (which contains the raw CRD files).

---

## Helm Chart Structure
The packaged Helm chart is located under `deploy/charts/jobcontroller/`:
```text
deploy/charts/jobcontroller/
├── Chart.yaml              # Chart versioning metadata
├── values.yaml             # Overridable values (e.g. image repository and tags)
├── crds/
│   └── crds.yaml           # Auto-compiled CRD manifests
└── templates/
    └── manifests.yaml      # Auto-compiled Deployment/RBAC manifests
```

### Values Configuration
Customize your deployment via `deploy/charts/jobcontroller/values.yaml`. 
For example, override the image settings:
```yaml
manager:
  image:
    repository: your-registry.azurecr.io/job-controller
    tag: v1.0.0
```

---

## Packaging and Publishing the Chart
To release and distribute the Helm chart to a production container registry (e.g. Azure Container Registry via OCI):

1. **Package the chart**:
   ```bash
   helm package deploy/charts/jobcontroller
   ```
   *This creates a tarball named `jobcontroller-<version>.tgz`.*
2. **Authenticate to your registry**:
   ```bash
   az acr login --name your-registry
   ```
3. **Push the OCI artifact**:
   ```bash
   helm push jobcontroller-<version>.tgz oci://your-registry.azurecr.io/helm
   ```

## Automatic Versioning & CI/CD Pipeline

The release pipeline automated via GitHub Actions (`.github/workflows/release.yml`) implements an automated versioning scheme:

1. **VERSION File**: The base version is specified in the [VERSION](file:///Users/david/Documents/Go/JobController/VERSION) file at the root of the repository (e.g., `0.1.0`).
2. **Dynamic Version Calculation**: 
   - The CI pipeline parses the `major.minor` components from the `VERSION` file.
   - It appends the GitHub Actions run number as the patch version.
   - For example, if the base version in `VERSION` is `0.1.0` and the GitHub Run Number is `4`, the calculated version is `0.1.4`.
3. **Synchronized Image and Chart Tags**:
   - The controller manager Docker image is built and pushed using this dynamic version tag (e.g., `0.1.4`), along with `latest` and the commit SHA tag.
   - The Helm chart is packaged using this exact same dynamic version for both the chart `version` and the `appVersion` (overriding the static version in `Chart.yaml`).
   - The default image repository in `values.yaml` is dynamically updated in the packaged chart to point to the correct registry destination.

---

## CLI Installation Guide

To deploy the operator from the remote OCI registry to your Kubernetes cluster:

### 1. Create target namespace
```bash
kubectl create namespace jobcontroller-system
```

### 2. Configure Pull Credentials
Because the controller manager image is hosted on a private repository, you must create a `docker-registry` secret named `jobcontroller-pull-secret` in the namespace where the operator runs:
```bash
kubectl create secret docker-registry jobcontroller-pull-secret \
  --docker-server=https://index.docker.io/v1/ \
  --docker-username="davidp0c" \
  --docker-password="YOUR_DOCKER_HUB_ACCESS_TOKEN_OR_PASSWORD" \
  --docker-email="your-email@example.com" \
  --namespace jobcontroller-system
```

### 3. Log in to the Helm OCI Registry
Authenticate your local Helm client to your Docker Hub registry:
```bash
echo "YOUR_DOCKER_HUB_PASSWORD" | helm registry login registry-1.docker.io --username davidp0c --password-stdin
```

### 4. Install/Upgrade the Operator via Helm
Deploy the operator using the remote OCI path and your specific release version:
```bash
helm upgrade --install jobcontroller oci://registry-1.docker.io/davidp0c/jobcontroller \
  --version <VERSION> \
  --namespace jobcontroller-system
```

---

## Deploying via Argo CD
Configure Argo CD to pull and deploy the operator directly.

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
    repoURL: 'https://dev.azure.com/dsandbox/Development/_git/KuberJobController'
    targetRevision: HEAD
    path: deploy/charts/jobcontroller
    helm:
      valueFiles:
        - values.yaml
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: job-controller-system
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
    repoURL: 'your-registry.azurecr.io/helm'
    targetRevision: 0.1.0  # Chart version
    helm:
      parameters:
        - name: manager.image.tag
          value: v1.0.0    # Controller image tag override
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: job-controller-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    createNamespace: true
```
