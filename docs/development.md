# Local Development Guide

This guide outlines how to set up your environment, implement features, write unit tests, and run the JobController operator locally.

## Prerequisites
Before you start developing, ensure you have the following tools installed locally:
- **Go** (version 1.25+ recommended)
- **Kubebuilder CLI** (version 4.14+)
- **Docker** (running container engine)
- **kubectl** (for interacting with clusters)

---

## Modifying API Schemes
The API schemas for the Custom Resources are defined in Go under the `api/` directory.

### Steps to Update API:
1. Open the schema definition file:
   [api/v1/jobrunner_types.go](file:///Users/david/Documents/Go/JobController/api/v1/jobrunner_types.go)
2. Add your desired fields to `JobRunnerSpec` (for configuration) or `JobRunnerStatus` (for tracking state). Remember to include appropriate `// +kubebuilder` markers for validation (e.g. `// +kubebuilder:validation:Required`).
3. Regenerate deepcopy helper functions and Kubernetes manifests (CRDs, RBAC rules) by running:
   ```bash
   make manifests generate
   ```

---

## Implementing Reconciliation Logic
Reconciliation logic resides in `internal/controller/jobrunner_controller.go`.

### Key Operator Rules:
- **Idempotency**: The `Reconcile` function must be safe to rerun multiple times. It should inspect the current cluster state and determine whether changes are needed.
- **Set Owner References**: When generating downstream resources (like a Kubernetes `Job` from a `JobRunner`), set the owner reference using `ctrl.SetControllerReference(jr, job, r.Scheme)`. This ensures cascading deletion when the parent CR is deleted.
- **Watch Downstream Resources**: Update `SetupWithManager` to watch resources owned by the controller so changes trigger reconciliation:
  ```go
  return ctrl.NewControllerManagedBy(mgr).
      For(&dtoolsv1.JobRunner{}).
      Owns(&batchv1.Job{}).
      Complete(r)
  ```

### Kubernetes Logging Guidelines
Please follow the standard Kubernetes logging guidelines when writing logs:
- Start message with a capital letter.
- Do not end messages with a period.
- Use past tense to represent actions that occurred (e.g. `"Created Job"`, not `"Create Job"`).
- Pass contextual key-value pairs (e.g. `"namespace"`, `"name"`).
  ```go
  log.Info("Created Job", "namespace", job.Namespace, "name", job.Name)
  log.Error(err, "Failed to get JobRunner", "name", req.Name)
  ```

---

## Running Tests
The unit tests utilize **Ginkgo + Gomega** and run against `envtest` (a local control plane simulating API Server and etcd, without starting full Pod execution).

### Running Unit Tests:
```bash
make test
```
The test suite automatically downloads the appropriate version of the local `envtest` binaries if they are not present in your local `./bin` directory.

---

## Running the Operator Locally
You can run the controller manager locally targeting your active `kubectl` cluster context:

1. **Install CRDs** to your cluster:
   ```bash
   make install
   ```
2. **Start the manager** process from your local machine:
   ```bash
   make run
   ```
   *Note: This will use your active Kubeconfig context (e.g., pointing to your development cluster).*
