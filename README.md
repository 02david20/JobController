# JobController

A Kubernetes operator built with Kubebuilder to run and manage container workloads using standard Kubernetes `batch/v1.Job` resources.

## Key Features

- **Custom Resource definition**: Deploys workloads defined by the `JobRunner` Custom Resource (under the `dtools.dsandbox.io/v1` API group).
- **Automated Workloads**: Manages the execution lifecycle of Kubernetes batch jobs from custom specifications.
- **Robust Status Tracking**: Reconciles the execution state (`Pending`, `Running`, `Completed`, `Failed`) and exposes it clearly in the resource's `.status.execute` field.
- **Decoupled Helm Chart**: Clean distribution via a Helm chart synced from Kustomize templates (`deploy/charts/jobcontroller`).
- **Dynamic CI/CD Pipeline**: Automates Docker image compilation and Helm chart packaging with synchronized version tagging (`major.minor.RUN_NUMBER`).

---

## Architecture & Resources

When a `JobRunner` custom resource is created, the operator reconciles it by:
1. Creating a corresponding `batch/v1.Job` matching the specified `image` and `command`.
2. Setting an OwnerReference on the Job to ensure garbage collection when the `JobRunner` is deleted.
3. Tracking the status of the underlying Job and updating the `JobRunner` status conditions accordingly.

### Sample Manifest

```yaml
apiVersion: dtools.dsandbox.io/v1
kind: JobRunner
metadata:
  name: sample-workload
  namespace: default
spec:
  image: busybox:latest
  command:
    - sh
    - -c
    - "echo 'Running job controller task...' && sleep 5 && echo 'Done!'"
```

---

## Quick Start & Development

### Prerequisites

- **Go**: `v1.25`
- **Docker**: `v17.03+`
- **Kubectl**: `v1.24+`
- **Kubernetes Cluster**: Access to a local (e.g. Kind) or remote cluster

### Local Development

1. **Run Tests**:
   ```bash
   make test
   ```
2. **Run Operator Locally** (uses current kubeconfig context):
   ```bash
   make install
   make run
   ```

For detailed setup, debugging, and testing commands, see [docs/development.md](file:///Users/david/Documents/Go/JobController/docs/development.md).

---

## Deployment & Production

### Syncing Kustomize to Helm
Whenever you make updates to RBAC rules, CRD markers, or controller deployment manifests, synchronize the changes to the Helm chart:
```bash
make sync-helm
```

### Install via Helm
Deploy the operator to your cluster:
```bash
helm upgrade --install jobcontroller deploy/charts/jobcontroller \
  --namespace jobcontroller-system \
  --create-namespace
```

For detailed production instructions, Helm parameters, and Argo CD configurations, see [docs/distribution.md](file:///Users/david/Documents/Go/JobController/docs/distribution.md).

---

## Versioning & Releases

This project uses an automated versioning setup integrated with GitHub Actions (`.github/workflows/release.yml`):
- **Base Version**: Managed in the [VERSION](file:///Users/david/Documents/Go/JobController/VERSION) file (e.g., `0.1.0`).
- **Dynamic Patch Suffix**: The CI workflow extracts `major.minor` and appends the GitHub Run Number as the patch version (e.g., `0.1.4`).
- **Releases**: Pushes matching Docker image tags and Helm OCI charts to Docker Hub under the generated version tag.

---

## License

Copyright 2026. Licensed under the Apache License, Version 2.0.
