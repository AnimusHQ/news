# Architecture Decisions

## ADR-001: Artifact-driven pipeline

Status: accepted.

Decision: every pipeline stage produces a typed artifact.

Rationale: generated media systems become impossible to debug if the only durable output is a video file. Typed artifacts make the system auditable, replayable, testable, and reviewable.

Consequences:

- more upfront schema design;
- easier QA and regression testing;
- better source provenance;
- safer automation.

## ADR-002: Multimodel architecture

Status: accepted.

Decision: the project must support multiple models and providers per task category.

Rationale: no single model should become the permanent authority. Different models perform better at different tasks: research synthesis, code reasoning, writing, safety review, visual reasoning, speech, and analytics.

Consequences:

- model registry required;
- task router required;
- provider adapters required;
- benchmark-driven selection required;
- stronger resilience and quality.

## ADR-003: Multimodel council before human QA

Status: accepted.

Decision: critical artifacts must be reviewed by multiple models before they reach the human operator.

Rationale: the operator should receive the best possible candidate, including approvals, objections, dissent, and unresolved risks.

Consequences:

- increased cost for high-risk stages;
- better detection of errors;
- reduced single-model bias;
- human QA remains final authority.

## ADR-004: Human authority for release

Status: accepted.

Decision: no generated content may be published publicly without human release approval.

Rationale: automated publishing can create reputational, factual, safety, or legal harm.

Consequences:

- release may be slower;
- trust is preserved;
- accountability is clear.

## ADR-005: Deterministic rendering where possible

Status: accepted.

Decision: explanatory visuals should prefer deterministic render systems over uncontrolled AI video generation.

Rationale: IT education needs accurate diagrams, code, terminal flows, and architecture visuals. Deterministic rendering improves consistency and correctness.

Preferred tools:

- Remotion;
- FFmpeg;
- SVG;
- Mermaid;
- Graphviz;
- Manim.

AI-generated visuals may be used only when they are provenance-tracked and do not mislead.

## ADR-006: Source hierarchy

Status: accepted.

Decision: primary sources outrank secondary sources; community signals are not authoritative by default.

Rationale: educational content must be technically precise.

Source order:

1. official docs;
2. source code;
3. standards / RFCs;
4. release notes;
5. maintainer statements;
6. reputable technical books/blogs;
7. community discussions as signal.

## ADR-007: Private/scheduled publishing first

Status: accepted.

Decision: publication flow must stage videos as private or scheduled before public release.

Rationale: final preview and metadata QA are required.

Consequences:

- direct public uploads are forbidden;
- release approval is explicit;
- corrections can happen before exposure.

## ADR-008: Analytics cannot override trust

Status: accepted.

Decision: analytics may influence topics, formats, and pacing, but cannot override editorial and verification gates.

Rationale: optimizing only for CTR or retention can degrade accuracy and trust.

Consequences:

- no misleading thumbnails;
- no sensationalism as default;
- quality remains the north star.

## ADR-009: Provider independence

Status: accepted.

Decision: core pipeline logic must not depend on one AI provider, TTS provider, or publishing adapter.

Rationale: models, pricing, policies, and capabilities change. The system must remain durable.

Consequences:

- provider adapter interfaces required;
- model capability registry required;
- fallback policies required;
- portability improves.

## ADR-010: Visible stylized mascot

Status: accepted.

Decision: the mascot should be visibly stylized rather than photorealistic.

Rationale: stylization reduces synthetic-realism risk, disclosure complexity, and impersonation concerns while improving brand distinctiveness.

Consequences:

- stronger brand identity;
- lower misleading media risk;
- easier animation pipeline.
