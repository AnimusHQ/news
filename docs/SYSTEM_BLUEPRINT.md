# Animus News System Blueprint

## 1. Purpose

Animus News is a production-grade, source-grounded, multimodel media system for creating high-quality educational IT content around the Animus open-source community.

The system is not an AI content farm. It is a **content compiler**: it transforms trusted sources, community knowledge, editorial intent, model-assisted reasoning, verification evidence, and production assets into technically precise educational media.

## 2. Non-negotiable principles

1. **No claim without a source.**
2. **No script without a research pack.**
3. **No final model monopoly.** Important decisions are reviewed by multiple models and then by a human operator.
4. **No render without technical verification.**
5. **No publication without QA.**
6. **No direct public publishing from generated output.**
7. **No reused content without meaningful transformation.**
8. **No silent failure.**
9. **Every episode must be replayable from typed artifacts.**
10. **Every optimization must preserve trust, safety, and educational value.**

## 3. Architecture at a glance

```mermaid
flowchart TD
  A[Trusted + Untrusted Sources] --> B[Ingestion + Normalization]
  B --> C[Topic Intelligence]
  C --> D[Research Pack]
  D --> E[Claim Graph]
  D --> F[Editorial Brief]
  F --> G[Script Draft]
  E --> H[Multimodel Verification Council]
  G --> H
  H --> I[Human Operator QA]
  I --> J[Storyboard]
  J --> K[Mascot + Visual Production]
  K --> L[Render Pipeline]
  L --> M[Production QA]
  M --> N[Scheduled Publish]
  N --> O[Analytics + Community Feedback]
  O --> C
```

## 4. Multimodel foundation

Animus News is model-agnostic. It must support multiple neural networks per task category and dynamically route work to the best available model for each task.

The goal is to avoid a single-model worldview, reduce hallucination risk, improve specialization, and preserve architectural independence from any one provider.

### 4.1 Model categories

```mermaid
flowchart TB
  subgraph ModelRegistry[Model Registry]
    R[Registered Models]
    C[Capabilities]
    B[Benchmarks]
    P[Provider Metadata]
    Cost[Cost / Latency / Context]
    Risk[Risk Profile]
  end

  subgraph TaskCategories[Task Categories]
    T1[Research Synthesis]
    T2[Technical Verification]
    T3[Script Writing]
    T4[Editorial Review]
    T5[Storyboard Planning]
    T6[Visual Reasoning]
    T7[Code / Diagram Generation]
    T8[Voice / TTS]
    T9[Policy + Safety Review]
    T10[Analytics Interpretation]
  end

  ModelRegistry --> Router[Task Router]
  TaskCategories --> Router
  Router --> Selected[Best Model Set per Task]
```

Each model record must include:

- provider;
- model identifier;
- modality support;
- context length;
- structured output reliability;
- tool-use reliability;
- reasoning strength;
- code strength;
- multilingual strength;
- safety behavior;
- latency;
- cost;
- privacy posture;
- supported deployment modes;
- benchmark history;
- known failure modes.

### 4.2 No single-model authority

Critical artifacts are reviewed by a **Multimodel Verification Council** before they reach the human operator.

```mermaid
sequenceDiagram
  participant WriterModel
  participant VerifierA
  participant VerifierB
  participant VerifierC
  participant Arbiter
  participant HumanQA

  WriterModel->>Arbiter: script.md + claims.json
  Arbiter->>VerifierA: verify technical correctness
  Arbiter->>VerifierB: verify clarity + completeness
  Arbiter->>VerifierC: verify safety + policy + bias
  VerifierA-->>Arbiter: approval / objections
  VerifierB-->>Arbiter: approval / objections
  VerifierC-->>Arbiter: approval / objections
  Arbiter->>Arbiter: reconcile consensus + disagreements
  Arbiter->>HumanQA: final candidate + dissent report
  HumanQA-->>Arbiter: approve / revise / block
```

The operator receives:

- final candidate;
- model approvals;
- dissenting opinions;
- unresolved risks;
- unsupported claims;
- quality score;
- recommended decision.

### 4.3 Consensus modes

Different artifacts require different approval policies.

| Artifact | Suggested model policy | Human gate |
|---|---:|---:|
| Topic shortlist | 2-of-3 advisory agreement | Required |
| Research pack | primary model + verifier model | Required for high-risk topics |
| Claims | strict verifier agreement | Required for disputed claims |
| Script | writer + editorial reviewer + technical reviewer | Required |
| Storyboard | creative model + visual reviewer | Optional for low-risk episodes |
| QA report | safety reviewer + technical reviewer + production reviewer | Required |
| Publish manifest | deterministic checks + release reviewer | Required |

### 4.4 Model routing

```mermaid
flowchart TD
  A[Task Request] --> B[Classify Task]
  B --> C[Load Model Candidates]
  C --> D[Filter by Capability]
  D --> E[Filter by Policy + Privacy]
  E --> F[Rank by Benchmarks]
  F --> G[Estimate Cost + Latency]
  G --> H{Critical Task?}
  H -->|Yes| I[Select Model Panel]
  H -->|No| J[Select Best Single Model]
  I --> K[Run Council]
  J --> L[Run Task]
  K --> M[Persist Outputs + Scores]
  L --> M
```

Model routing must be configurable, benchmark-driven, and reversible. No application code should hard-code one provider as the permanent authority.

## 5. End-to-end episode lifecycle

```mermaid
stateDiagram-v2
  [*] --> Candidate
  Candidate --> ApprovedTopic: editor approval
  ApprovedTopic --> Researching
  Researching --> ResearchReady: research_pack accepted
  ResearchReady --> Drafting
  Drafting --> Verification
  Verification --> Revision: objections found
  Revision --> Verification
  Verification --> HumanQA: multimodel council approves
  HumanQA --> Storyboarding: operator approves
  HumanQA --> Blocked: operator blocks
  Storyboarding --> AssetProduction
  AssetProduction --> Rendering
  Rendering --> ProductionQA
  ProductionQA --> FixRequired: QA failed
  FixRequired --> Rendering
  ProductionQA --> Scheduled: QA approved
  Scheduled --> Published
  Published --> Monitored
  Monitored --> Archived
```

## 6. Canonical artifact graph

```mermaid
flowchart TD
  T[topic.yaml] --> RP[research_pack.json]
  RP --> CL[claims.json]
  RP --> EB[editorial_brief.md]
  EB --> SC[script.md]
  CL --> VR[verification_report.json]
  SC --> VR
  VR --> MC[multimodel_approval_report.json]
  MC --> HQ[human_qa_report.json]
  HQ --> SB[storyboard.yaml]
  SB --> AM[asset_manifest.json]
  AM --> RM[render_manifest.json]
  RM --> PQ[production_qa_report.json]
  PQ --> PM[publish_manifest.json]
  PM --> AR[analytics_report.json]
```

Episode directory:

```text
episodes/<episode-id>/
  topic.yaml
  research_pack.json
  claims.json
  editorial_brief.md
  script.md
  verification_report.json
  multimodel_approval_report.json
  human_qa_report.json
  storyboard.yaml
  asset_manifest.json
  render_manifest.json
  production_qa_report.json
  publish_manifest.json
  analytics_report.json
```

## 7. Knowledge layer

Sources are separated by trust level.

```mermaid
flowchart TB
  subgraph Primary[Primary Sources]
    P1[Official Documentation]
    P2[Specs / RFCs]
    P3[Source Code]
    P4[Release Notes]
    P5[Maintainer Statements]
  end

  subgraph Secondary[Secondary Sources]
    S1[Engineering Blogs]
    S2[Conference Talks]
    S3[Books]
    S4[Reputable Technical Articles]
  end

  subgraph Community[Community Signals]
    C1[GitHub Issues]
    C2[Discussions]
    C3[Comments]
    C4[Questions]
  end

  Primary --> Rank[Source Ranking]
  Secondary --> Rank
  Community --> Rank
  Rank --> Normalize[Normalize + Hash + Store]
  Normalize --> Research[Research Pack Builder]
```

Primary sources outrank secondary sources. Community signals are useful for topic selection and examples, but they do not become authoritative evidence without verification.

## 8. Research pack builder

```mermaid
flowchart TD
  A[Approved Topic] --> B[Source Discovery]
  B --> C[Source Deduplication]
  C --> D[Trust Ranking]
  D --> E[Source Extraction]
  E --> F[Claim Candidate Extraction]
  F --> G[Contradiction Scan]
  G --> H[Terminology Map]
  H --> I[Forbidden Simplifications]
  I --> J[Research Pack]
```

The research pack must define:

- core question;
- target audience;
- learning objectives;
- trusted sources;
- extracted claims;
- unresolved questions;
- known controversies;
- required terminology;
- forbidden simplifications;
- visual opportunities;
- CTA alignment.

## 9. Claim graph

```mermaid
classDiagram
  class Episode {
    string id
    string title
    string format
    string status
  }
  class Claim {
    string id
    string text
    string type
    string risk_level
    string status
  }
  class Source {
    string id
    string uri
    string trust_level
    string content_hash
    string license_notes
  }
  class EvidenceLocator {
    string source_id
    string section
    string range
    string quote_hash
  }
  class ModelReview {
    string model_id
    string verdict
    float confidence
    string notes
  }

  Episode "1" --> "many" Claim
  Claim "many" --> "many" Source
  Claim "many" --> "many" EvidenceLocator
  Claim "many" --> "many" ModelReview
```

Claim statuses:

- `supported`
- `partially_supported`
- `unsupported`
- `contradicted`
- `needs_human_review`
- `removed`

No high-risk claim may proceed with `unsupported`, `contradicted`, or `needs_human_review` status.

## 10. Production layer

The production layer converts approved script and storyboard into media assets.

```mermaid
flowchart LR
  A[Approved Storyboard] --> B[Mascot Director]
  A --> C[Visual Engine]
  A --> D[Caption Engine]
  B --> E[Performance Manifest]
  C --> F[Diagram + Code + UI Assets]
  D --> G[Subtitles]
  E --> H[Render Engine]
  F --> H
  G --> H
  H --> I[Master Render]
  I --> J[YouTube 16:9]
  I --> K[Shorts 9:16]
  I --> L[Thumbnail Candidates]
  I --> M[Transcript]
```

Rendering should prefer deterministic and inspectable assets:

- Remotion / React video for layout;
- FFmpeg for encoding;
- SVG/Mermaid/Graphviz for diagrams;
- Manim for technical animations;
- controlled mascot rig;
- content-addressed asset cache;
- explicit asset manifests.

## 11. Publishing layer

Publishing is staged and gated.

```mermaid
flowchart TD
  A[Render Complete] --> B[Production QA]
  B --> C{Approved?}
  C -->|No| D[Fix Required]
  D --> A
  C -->|Yes| E[Private Upload]
  E --> F[Metadata QA]
  F --> G{Release Approved?}
  G -->|No| H[Blocked]
  G -->|Yes| I[Scheduled Publish]
  I --> J[Published]
  J --> K[Monitoring]
```

Direct public publishing from generation is forbidden.

## 12. Analytics loop

Analytics improve future content but must not degrade trust.

```mermaid
flowchart TD
  A[Published Episode] --> B[Metric Import]
  B --> C[CTR Analysis]
  B --> D[Retention Analysis]
  B --> E[Comment Mining]
  B --> F[Community Conversion]
  C --> G[Insight Report]
  D --> G
  E --> G
  F --> G
  G --> H[Topic Scoring Updates]
  G --> I[Format Improvements]
  G --> J[Mascot / Visual Improvements]
```

The system should optimize for:

- clarity;
- retention;
- trust;
- community conversion;
- technical accuracy;
- production efficiency.

It must not optimize for misleading clickbait, sensationalism, or engagement bait.

## 13. Deployment architecture

```mermaid
flowchart TB
  UI[Editorial Console] --> API[Backend API]
  API --> AUTH[Auth / RBAC]
  API --> WF[Workflow Engine]
  WF --> DB[(Postgres)]
  WF --> OBJ[(Object Storage)]
  WF --> SEARCH[(Hybrid Search Index)]
  WF --> Q[Task Queue]
  WF --> MR[Model Registry]
  MR --> ROUTER[Model Router]
  ROUTER --> LLM1[LLM Provider A]
  ROUTER --> LLM2[LLM Provider B]
  ROUTER --> LLM3[LLM Provider C]
  ROUTER --> VISION[Vision Models]
  ROUTER --> TTS[TTS Providers]
  WF --> RENDER[Render Workers]
  RENDER --> OBJ
  WF --> PUB[Publishing Worker]
  PUB --> YT[YouTube]
  API --> OBS[Observability]
```

## 14. Trust boundaries

```mermaid
flowchart TB
  subgraph Untrusted[Untrusted Inputs]
    U1[Web Pages]
    U2[Comments]
    U3[Community Submissions]
    U4[Generated Media]
    U5[Model Outputs]
  end

  subgraph Controlled[Controlled Processing]
    P1[Sandboxed Ingestion]
    P2[Normalization]
    P3[Source Ranking]
    P4[Claim Extraction]
    P5[Multimodel Verification]
    P6[Human QA]
  end

  subgraph Trusted[Release Artifacts]
    T1[Approved Script]
    T2[QA Reports]
    T3[Publish Manifest]
  end

  Untrusted --> P1 --> P2 --> P3 --> P4 --> P5 --> P6 --> Trusted
```

Model outputs are untrusted until validated. A model approval is evidence, not authority.

## 15. System success criteria

Animus News succeeds when it can repeatedly produce episodes that are:

- technically accurate;
- source-grounded;
- visually clear;
- original and transformative;
- useful to newcomers;
- respected by experienced engineers;
- aligned with the Animus community;
- safe to scale;
- auditable after publication;
- efficient enough for low manual operator time;
- independent from any single AI provider.
