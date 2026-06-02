<h1 align="center">Grewanderer</h1>

<p align="center">
  <strong>Go/Rust Platform Engineer for secure device-connected infrastructure</strong><br/>
  Control planes · Management planes · Secure service exposure · Embedded Linux · Reliability · Observability
</p>

<p align="center">
  <a href="mailto:rewanderer@proton.me">Email</a>
  ·
  <a href="https://kapakka.org">Website</a>
  ·
  <a href="https://github.com/grewanderer/animus-link">Animus Link</a>
  ·
  <a href="https://github.com/grewanderer/animus_link_manager">Animus Link Manager</a>
  ·
  <a href="https://github.com/grewanderer/animus-amity">Amity</a>
  ·
  <a href="https://github.com/AnimusHQ">AnimusHQ</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-management%20plane-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go management plane" />
  <img src="https://img.shields.io/badge/Rust-control%20plane-000000?style=flat-square&logo=rust&logoColor=white" alt="Rust control plane" />
  <img src="https://img.shields.io/badge/Linux-device%20infrastructure-FCC624?style=flat-square&logo=linux&logoColor=black" alt="Linux device infrastructure" />
  <img src="https://img.shields.io/badge/Kubernetes-platform-326CE5?style=flat-square&logo=kubernetes&logoColor=white" alt="Kubernetes platform" />
  <img src="https://img.shields.io/badge/OpenTelemetry-observability-000000?style=flat-square&logo=opentelemetry&logoColor=white" alt="OpenTelemetry observability" />
</p>

---

## What I build

I build backend, platform, and systems software for infrastructure where software has to coordinate real devices, private services, secure access paths, operational state, and production failure modes.

My strongest work is at the boundary between **Go management-plane services**, **Rust control-plane / systems components**, and **Linux-based device or infrastructure environments**.

I care about systems that remain understandable after they ship: explicit state ownership, reproducible validation, observable runtime behavior, predictable failure semantics, and security boundaries that are concrete enough to test.

---

## Current focus

- **Secure device and service access**: relay-first access, private service exposure, identity-aware sessions, controlled connectivity.
- **Control-plane engineering**: typed command models, state machines, validation, runtime effect checks, and deterministic operational behavior.
- **Management-plane backends**: Go APIs, device and service registries, session orchestration, audit logs, health/status surfaces, operator workflows.
- **Embedded and network-appliance validation**: Buildroot, QEMU, Linux images, CI smoke checks, release gates, firmware-oriented regression safety.
- **Reliability and observability**: logs, metrics, traces, audit records, incident/debug paths, Prometheus, Grafana, OpenTelemetry.

---

## Selected systems

### [Animus Link](https://github.com/grewanderer/animus-link)

Rust-based secure connectivity and service-exposure substrate exploring relay-first access for private infrastructure.

What it demonstrates:

- separation between identity, transport, relay, session, and service-exposure boundaries;
- relay-assisted connectivity across constrained or untrusted networks;
- end-to-end encrypted session direction where relays are not positioned as payload owners;
- invite-first private discovery and controlled service exposure;
- conformance-oriented thinking around protocol behavior and runtime expectations;
- a foundation for secure device-connected infrastructure rather than ad-hoc tunnels.

Status: early-stage open-source system design and implementation work. It is a technical proof of direction, not a claim of production security certification.

### [Animus Link Manager](https://github.com/grewanderer/animus_link_manager)

Go-based management-plane layer around secure device/service access and operational workflows.

What it is intended to cover:

- service and device inventory;
- identity/session management;
- relay and exposure configuration;
- health, status, and audit surfaces;
- operator-facing workflows;
- API-first management around lower-level connectivity primitives.

This project represents the operational side of the Animus Link direction: the part that turns secure connectivity primitives into something a team can run, inspect, and manage.

### [Amity](https://github.com/grewanderer/animus-amity)

Document-governed AI-assisted software delivery system focused on controlled implementation, verification, and execution evidence.

Relevant ideas:

- explicit role separation between architecture, verification, and execution;
- deterministic approval and bounded execution scopes;
- typed artifacts for designs, reviews, implementation evidence, and completion records;
- durable recovery from partial failures, worker crashes, rate limits, and incomplete execution;
- provenance-first retrieval and evidence handling rather than blind automation.

Amity is secondary to my device-infrastructure work, but it reflects the same engineering pattern: controlled state, explicit evidence, reproducible execution, and operational discipline.

### Non-public industrial / network-appliance work

Some of my strongest engineering work is not public. The public-safe description is:

- Rust-based control-plane logic for a network appliance with strict reliability and validation requirements;
- Go/Rust backend components for distributed industrial systems;
- backend integration with embedded devices and platform-specific components;
- Buildroot/QEMU-based validation paths for firmware-oriented workflows;
- secure transport, operational diagnostics, CI/CD, and production support for infrastructure-adjacent systems.

I do not publish confidential implementation details, employer code, private protocol decisions, or customer-specific architecture.

---

## Engineering surface

| Area | What I build | Technical focus |
|---|---|---|
| Management planes | APIs and operational backends for devices, services, sessions, and operators | Go, REST, gRPC, PostgreSQL, audit logs, health/status, service boundaries |
| Control planes | command/state logic, runtime behavior, validation and effect modeling | Rust, Go, state machines, typed commands, deterministic behavior, conformance checks |
| Secure access | controlled service exposure and relay-first connectivity | identity, sessions, TLS/DTLS, secure transport, policy boundaries, private discovery |
| Embedded Linux | Linux-based device/platform environments and validation paths | Buildroot, QEMU, rootfs composition, firmware delivery, package integration |
| Platform engineering | delivery, deployment, CI/CD, release gates, reproducibility | Linux, Docker, Kubernetes, Helm, GitHub Actions, GitLab CI, Jenkins |
| Reliability | systems that can be debugged under real production constraints | Prometheus, Grafana, OpenTelemetry, structured logs, metrics, traces, incident paths |
| Integration-heavy backend | backend services around protocols, devices, and constrained environments | Go, Rust, Python, REST, gRPC, WebSockets, queues, distributed workflows |

---

## Technical stack

```text
Languages:        Go, Rust, Python
Backend:          REST, gRPC, WebSockets, event-driven systems, message-driven systems
Data:             PostgreSQL, MySQL, SQLite, Redis, Kafka, RabbitMQ, S3-compatible storage
Platform:         Linux, Docker, Docker Compose, Kubernetes, Helm, Kustomize, Buildroot, QEMU
Delivery:         GitHub Actions, GitLab CI, Jenkins, reproducible builds, release gates
Observability:    Prometheus, Grafana, OpenTelemetry, structured logs, metrics, traces, ELK
Security:         TLS, DTLS, OIDC, JWT, RBAC, ACLs, policy enforcement, secure transport
Systems focus:    control planes, management planes, identity/session models, service exposure
```

---

## How I think about engineering

- Runtime behavior matters more than repository aesthetics.
- Parser support is not feature support; runtime effect and readback matter.
- Control planes should own truth; execution environments should report evidence.
- Production systems need rollback, observability, state ownership, and defined failure semantics.
- Logs, metrics, traces, and audit records are design surfaces.
- Reproducibility is a debugging and accountability mechanism, not a tooling preference.
- Security boundaries should be explicit enough to test and boring enough to operate.
- Good infrastructure makes failure visible before it becomes an incident.

---

## Good fit

I am a strong fit for teams building:

- secure access to private devices, internal services, labs, or edge infrastructure;
- Go management planes around devices, sessions, services, operators, and audit trails;
- Rust control-plane or systems components for network/industrial appliances;
- embedded Linux release and validation pipelines;
- reliability-critical backend systems connected to real-world infrastructure;
- platform services where correctness, observability, and operational clarity matter.

Relevant collaboration formats:

- fixed-scope architecture review;
- control-plane / management-plane implementation;
- secure device-access prototype;
- embedded Linux validation pipeline;
- reliability and observability audit;
- technical due diligence for infrastructure-heavy products.

---

<p align="center">
  <sub>I build systems that keep private infrastructure accessible, observable, reproducible, and operationally understandable after the first implementation has shipped.</sub>
</p>
