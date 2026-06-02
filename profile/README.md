<h1 align="center">AnimusHQ</h1>

<p align="center">
  <strong>Open-source systems for secure, observable, and reproducible device-connected infrastructure.</strong><br/>
  Relay-first access · Control planes · Management planes · Embedded Linux validation · Reliability engineering
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

## Mission

AnimusHQ builds open-source infrastructure for systems where private services, real devices, secure access paths, operational state, and production failure modes meet.

The organization is focused on the engineering layer between low-level connectivity and usable operations:

- secure access to private services and device-connected environments;
- relay-first connectivity across constrained networks;
- identity-aware sessions and controlled service exposure;
- management-plane APIs and operator workflows;
- control-plane validation, conformance, and runtime evidence;
- reproducible delivery, observability, and release safety.

The goal is not to create another ad-hoc tunneling tool. The goal is to build understandable infrastructure that can be inspected, validated, operated, and evolved.

---

## Core direction

### Secure device and service access

AnimusHQ explores relay-first service exposure for environments where direct connectivity is unreliable, unsafe, or operationally expensive.

Key concerns:

- private service exposure without broad network access;
- identity-aware sessions instead of anonymous tunnels;
- explicit service registration and authorization boundaries;
- relay participation without treating the relay as the owner of payload semantics;
- auditability and operator visibility as first-class design requirements.

### Control-plane and management-plane systems

The organization separates infrastructure into two layers:

- **Control plane**: core behavior, state transitions, command/session semantics, validation, conformance, and runtime rules.
- **Management plane**: APIs, service/device inventory, health, audit logs, operator workflows, configuration, and operational policy.

This separation keeps low-level correctness independent from operator-facing workflows while still making the system manageable in production.

### Reproducible operations

AnimusHQ treats logs, metrics, traces, audit records, validation reports, and release artifacts as part of the product surface.

A system is not considered production-ready only because it runs. It must also be diagnosable, reproducible, observable, and safe to change.

---

## Projects

### [Animus Link](https://github.com/AnimusHQ/link)

Relay-first secure access substrate for private services and device-connected infrastructure.

Primary engineering themes:

- identity, session, relay, and service-exposure boundaries;
- secure transport and controlled connectivity;
- private discovery and explicit authorization paths;
- conformance-oriented protocol behavior;
- reproducible runtime validation.

### [Animus Link UX](https://github.com/AnimusHQ/link_ux)

Operator-facing user experience for managing secure access paths, exposed services, session state, and system visibility.

Primary engineering themes:

- service and device inventory;
- session and access visualization;
- health/status surfaces;
- operator workflows;
- configuration and auditability.

### [Animus News](https://github.com/AnimusHQ/news)

Source-grounded technical media and knowledge system for documenting engineering practice, open-source systems, and infrastructure concepts.

This project is secondary to the infrastructure direction. It exists to support trustworthy technical communication, not to dilute the core engineering focus.

---

## Engineering principles

- Runtime behavior matters more than repository aesthetics.
- Security boundaries should be explicit enough to test and boring enough to operate.
- Control planes should own truth; execution environments should report evidence.
- Management planes should make state visible, inspectable, and auditable.
- Logs, metrics, traces, and audit records are design surfaces.
- Reproducibility is a debugging and accountability mechanism, not a tooling preference.
- Parser support is not feature support; runtime effect and readback matter.
- Production systems need rollback, observability, state ownership, and defined failure semantics.

---

## Production-grade bar

AnimusHQ projects should move toward the following baseline before being presented as production-ready:

- clear README with scope, status, non-goals, and threat model where relevant;
- explicit license and contribution policy;
- reproducible local development path;
- automated formatting, linting, tests, and security checks;
- CI release gates and artifact validation;
- documented architecture decisions;
- observability and operational runbooks;
- issue templates, pull request templates, and security reporting path;
- versioned releases and changelog once external users depend on the project.

Until a project meets this bar, it should be described as an early-stage prototype, research implementation, or engineering proof rather than as a production-certified system.

---

## Collaboration

AnimusHQ is currently most relevant for teams working on:

- secure access to private devices, labs, internal services, or edge infrastructure;
- Go management planes around devices, sessions, services, operators, and audit trails;
- Rust control-plane components for network or industrial appliances;
- embedded Linux validation and release-safety workflows;
- reliability-critical backend systems connected to real-world infrastructure.

For collaboration, architecture review, or technical discussion: [rewanderer@proton.me](mailto:rewanderer@proton.me)

---

<p align="center">
  <sub>AnimusHQ builds infrastructure that stays accessible, observable, reproducible, and operationally understandable after the first implementation has shipped.</sub>
</p>
