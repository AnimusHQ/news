<h1 align="center">AnimusHQ</h1>

<p align="center">
  <strong>Secure access and control-plane systems for device-connected infrastructure.</strong><br/>
  Relay-first connectivity · Service exposure · Management planes · Embedded Linux validation · Runtime observability
</p>

<p align="center">
  <a href="mailto:rewanderer@proton.me">Contact</a>
  ·
  <a href="https://github.com/AnimusHQ/link">Animus Link</a>
  ·
  <a href="https://github.com/AnimusHQ/link_ux">Animus Link UX</a>
  ·
  <a href="https://github.com/AnimusHQ/news">Animus News</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Rust-control%20plane-000000?style=flat-square&logo=rust&logoColor=white" alt="Rust control plane" />
  <img src="https://img.shields.io/badge/Go-management%20plane-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go management plane" />
  <img src="https://img.shields.io/badge/Linux-device%20infrastructure-FCC624?style=flat-square&logo=linux&logoColor=black" alt="Linux device infrastructure" />
  <img src="https://img.shields.io/badge/OpenTelemetry-observability-000000?style=flat-square&logo=opentelemetry&logoColor=white" alt="OpenTelemetry observability" />
</p>

---

## Scope

AnimusHQ builds infrastructure components for private services, device-connected systems, and constrained networks.

The organization focuses on five engineering areas:

1. **Relay-first access** — exposing private services without assuming direct inbound connectivity.
2. **Identity-aware sessions** — binding access to explicit identities, sessions, policies, and audit records.
3. **Control-plane correctness** — validating state transitions, command semantics, protocol behavior, and runtime effects.
4. **Management-plane operations** — APIs, inventories, health views, configuration surfaces, and operator workflows.
5. **Release safety** — reproducible builds, CI gates, conformance checks, logs, metrics, traces, and validation artifacts.

AnimusHQ does not publish production-security claims until a project has a documented threat model, security policy, supported configuration, release process, and vulnerability handling path.

---

## Project status

| Project | Purpose | Public status |
|---|---|---|
| [Animus Link](https://github.com/AnimusHQ/link) | Relay-first secure access substrate for private services and device-connected infrastructure. | Active prototype / architecture implementation. Not security-certified. |
| [Animus Link UX](https://github.com/AnimusHQ/link_ux) | Operator interface for exposed services, sessions, health, access state, and configuration. | Early interface layer. Not a supported admin console yet. |
| [Animus News](https://github.com/AnimusHQ/news) | Source-grounded technical media and documentation workflow system. | Secondary project. Not part of the secure-access runtime path. |

---

## Architecture model

AnimusHQ projects use a two-layer operating model.

### Control plane

The control plane owns runtime semantics:

- identity and session state;
- command validation;
- service exposure rules;
- relay and transport behavior;
- conformance checks;
- runtime evidence.

### Management plane

The management plane owns operator workflows:

- service and device inventory;
- configuration APIs;
- health and status views;
- audit logs;
- access/session visibility;
- release and operational diagnostics.

This separation prevents UI/API workflows from becoming the source of truth for low-level behavior.

---

## Non-goals

AnimusHQ is not currently positioning its projects as:

- a drop-in VPN replacement;
- a production-certified zero-trust platform;
- a managed cloud service;
- a security product with formal assurance claims;
- a generic tunneling utility;
- a content or media company.

The current public work is an engineering foundation for secure device-connected infrastructure. Production claims require release discipline, security review, supported deployments, and documented operations.

---

## Engineering standards

All production-facing AnimusHQ repositories should define:

- project scope, status, and non-goals;
- supported local development path;
- license and contribution rules;
- security policy and vulnerability reporting path;
- architecture notes or ADRs;
- formatting, linting, tests, and CI gates;
- threat model when security boundaries are involved;
- observability surfaces for runtime systems;
- release process, changelog, and versioning once users depend on the project.

A repository that does not meet this baseline must describe itself as an early-stage prototype, research implementation, or engineering proof.

---

## Operating principles

- Runtime behavior is part of the API.
- Security boundaries must be testable.
- Control planes own truth; management planes expose and operate it.
- Logs, metrics, traces, audit records, and validation reports are product surfaces.
- Reproducibility is required for debugging, accountability, and release safety.
- Parser support is not feature support; runtime effect and readback matter.
- Production readiness requires rollback, observability, ownership of state, and documented failure semantics.

---

## Collaboration

AnimusHQ is relevant to teams working on:

- secure access to private services, internal tools, labs, edge systems, or device fleets;
- Go management planes for devices, sessions, services, operators, and audit trails;
- Rust control-plane components for network or industrial appliances;
- embedded Linux validation and release-safety workflows;
- reliability-critical backend systems connected to physical or private infrastructure.

Contact: [rewanderer@proton.me](mailto:rewanderer@proton.me)

---

<p align="center">
  <sub>AnimusHQ builds infrastructure that remains accessible, observable, reproducible, and operationally understandable after the first implementation ships.</sub>
</p>
