# 📚 Available Content Directory

Welcome to the documentation directory for the **JobController Operator**. Use this page to quickly locate guides, setup templates, and deployment references.

---

## 🗂️ Documentation Sections

<div class="card-grid">

  <!-- Card 1: Home/Overview -->
  <div class="card-item">
    <h3>🏠 <a href="../index.md">Home & Overview</a></h3>
    <p>The entry point for the operator, introducing key custom resources and high-level architectural design concepts.</p>
    <hr>
    <ul>
      <li><a href="../index.md#overview">Operator Overview</a></li>
      <li><a href="../index.md#custom-resources">Custom Resources (JobRunner)</a></li>
      <li><a href="../index.md#architecture-design">Decoupled Helm/Kustomize Design</a></li>
      <li><a href="../index.md#project-structure">Project Directory Tree</a></li>
    </ul>
  </div>

  <!-- Card 2: Development Guide -->
  <div class="card-item">
    <h3>💻 <a href="../development.md">Development Guide</a></h3>
    <p>Step-by-step instructions on setting up local environments, modifying API schemes, writing reconcilers, and testing.</p>
    <hr>
    <ul>
      <li><a href="../development.md#prerequisites">Development Prerequisites</a></li>
      <li><a href="../development.md#modifying-api-schemes">Modifying API Schemes & Kubebuilder Markers</a></li>
      <li><a href="../development.md#implementing-reconciliation-logic">Reconciliation Rules & Kubernetes Logging</a></li>
      <li><a href="../development.md#running-tests">Running envtest Unit Tests</a></li>
      <li><a href="../development.md#running-the-operator-locally">Local Running (make run)</a></li>
    </ul>
  </div>

  <!-- Card 3: Helm Distribution -->
  <div class="card-item">
    <h3>☸️ <a href="../distribution.md">Helm Distribution</a></h3>
    <p>Guide to packaging the operator, configuration parameter tables, Makefile targets, and OCI distribution.</p>
    <hr>
    <ul>
      <li><a href="../distribution.md#1-architecture-kubebuilder-helm-plugin-helmv2-alpha">Kubebuilder Helm Plugin Setup</a></li>
      <li><a href="../distribution.md#2-valuesyaml-key-parameters">Exposed parameters in values.yaml</a></li>
      <li><a href="../distribution.md#3-development-management-makefile-targets">Syncing Kustomize, Linting & E2E Testing</a></li>
      <li><a href="../distribution.md#4-automatic-versioning-cicd-pipeline">GitHub Actions Release Pipeline</a></li>
      <li><a href="../distribution.md#5-cli-installation-guide">Command-Line Install from OCI</a></li>
      <li><a href="../distribution.md#6-gitops-deployment-via-argo-cd">Argo CD Deployment Specs</a></li>
    </ul>
  </div>

  <!-- Card 4: Setup Template -->
  <div class="card-item">
    <h3>🛠️ <a href="../SETUP_TEMPLATE.md">Setup Template</a></h3>
    <p>A reusable, end-to-end blueprint showing how to build similar operators and Helm distributions from scratch.</p>
    <hr>
    <ul>
      <li><a href="../SETUP_TEMPLATE.md#0-quick-start">10-Minute Quick Start Commands</a></li>
      <li><a href="../SETUP_TEMPLATE.md#2-initialise-the-project">Project Initialization</a></li>
      <li><a href="../SETUP_TEMPLATE.md#5-helm-distribution">Helm Configuration & Snapshot Tests</a></li>
      <li><a href="../SETUP_TEMPLATE.md#7-local-testing-with-kind">Kind Inner-Loop Workflows</a></li>
      <li><a href="../SETUP_TEMPLATE.md#8-cicd--github-actions-pipeline">Full GitHub Actions YAML Specs</a></li>
    </ul>
  </div>

</div>

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
