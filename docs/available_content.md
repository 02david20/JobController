# 📚 Available Content Directory

Welcome to the documentation directory for the **JobController Operator**. Use this page to quickly locate guides, setup templates, and deployment references.

---

## 🗂️ Documentation Sections

{{ pagenav() }}

---

## 🎯 Quick Reference Cheat Sheet

For a high-level summary of commands and files, here are the most frequently used paths and commands:

| Resource / Action | File / Command | Description |
| :--- | :--- | :--- |
| **API Custom Resource Specs** | [api/v1/jobrunner_types.go](file:///Users/david/Documents/Go/JobController/api/v1/jobrunner_types.go) | Define properties and status conditions of `JobRunner` |
| **Reconciliation Controller** | [internal/controller/jobrunner_controller.go](file:///Users/david/Documents/Go/JobController/internal/controller/jobrunner_controller.go) | Core loop processing `JobRunner` state |
| **Helm Values Configuration** | [dist/chart/values.yaml](file:///Users/david/Documents/Go/JobController/dist/chart/values.yaml) | Deployment customization for Helm |
| **Regenerate Manifests** | `make generate-dist` | Re-syncs Kustomize manifests, generates Helm chart templates, schemas, and lints |
| **Execute Unit Tests** | `make test` | Runs Ginkgo specs against envtest |
| **Run E2E Integrations** | `make test-e2e` / `make test-helm-e2e` | Builds image and runs tests inside a local Kind cluster |

> [!TIP]
> Use the sidebar navigation on the left of the documentation page to browse these files interactively.
