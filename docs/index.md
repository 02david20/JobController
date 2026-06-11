# JobController Operator

Welcome to the JobController Operator documentation. This operator is designed to manage long-running tasks or batch workloads inside Kubernetes using the `JobRunner` Custom Resource Definition (CRD).

## Overview
The JobController Operator monitors custom `JobRunner` resources in your Kubernetes cluster and coordinates their lifecycle by spinning up downstream Kubernetes `batch/v1.Job` workloads. It translates the success/failure state of those workloads back into the Custom Resource's status.

### Custom Resources
- **JobRunner**: Defines the workload specification (container image, command, and arguments) and tracks execution progress (e.g., `Pending`, `Running`, `Completed`, `Failed`).

### Architecture Design
To maintain enterprise-grade standards while keeping deployments lean, the operator uses a decoupled **Kustomize + Helm** distribution pipeline. 
- **Development**: The schema (`api/v1/jobrunner_types.go`) and reconciliation loops are designed using Kubebuilder and tested locally with `envtest`.
- **Distribution**: Code is compiled and packaged using Kustomize, which is then directly output to a single manifest inside the Helm chart templates. Any unused telemetry, Prometheus `ServiceMonitor`, or metrics-server patches are entirely excluded from the production-ready Helm chart, ensuring the footprint on production clusters is as minimal as possible.

## Project Structure
The repository is structured as follows:
```text
├── api/v1/                    # Go API definitions for the Custom Resource (CRD)
├── config/                    # Kustomize base templates for CRDs, RBAC, and manager deployment
├── deploy/charts/             # Production Helm chart (jobcontroller)
│   └── jobcontroller/
│       ├── crds/              # Compiles raw CRD manifests
│       └── templates/         # Compiles manager deployment + RBAC manifests
├── internal/controller/       # Reconciliation controller logic (JobRunnerReconciler)
└── cmd/main.go                # Manager entrypoint
```

---

## Next Steps
- Learn how to build and test the operator in the [Development Guide](development.md).
- Understand how the Helm chart is built and deployed in the [Helm Distribution Guide](distribution.md).
