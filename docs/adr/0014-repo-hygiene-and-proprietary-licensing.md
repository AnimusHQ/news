# ADR-0014: Repository Hygiene, Proprietary Licensing, and Vendor-Name Decoupling

Status: accepted.

## Context

REPO_HYGIENE_001 addressed documentation, licensing, and naming gaps without
changing behavior. Four decisions were non-trivial and are recorded here so a
fresh agent can reconstruct what changed and why.

1. The root `README.md` described "AnimusHQ organization defaults" and pointed at
   `profile/README.md`. That is org-template content, not a project README, and
   `profile/README.md` only renders as the GitHub org profile in a repository
   literally named `.github`. In `AnimusHQ/news` it was inert and misleading.
2. The repository is public but had no `LICENSE` and no explicit licensing
   status, leaving its status ambiguous and at risk of being read as open source.
3. Two design docs (`docs/SYSTEM_BLUEPRINT.md`, `docs/MULTIMODEL_STRATEGY.md`, and
   related forward-looking docs) describe the full target system as if built,
   while the implemented surface is the short-form M1–L2 slice on mock /
   fail-closed providers.
4. Short-form workflow activities embedded specific commercial vendor names
   (ElevenLabs, Upload-Post, Seedance) in the deterministic workflow layer, in
   tension with the provider-independence principle (`AGENTS.md`,
   `docs/MULTIMODEL_STRATEGY.md` §9).

## Decision

**Licensing.** The project is proprietary; all rights are retained exclusively by
Animus. A top-level `LICENSE` states all-rights-reserved and that public
visibility grants no license, scoped to this repository only (separately
published Animus open-source community projects keep their own terms). No SPDX
OSS identifier or permissive license text is added anywhere. `README.md`,
`CONTRIBUTING.md`, `SECURITY.md`, and `SUPPORT.md` are reconciled so no file
implies an open-source grant. `CONTRIBUTING.md` states that external
contributions are not accepted by default and, when invited, only under a written
agreement assigning rights to Animus.

**`profile/` handling.** `profile/README.md` is removed. Org-level GitHub profile
content belongs in `AnimusHQ/.github`, not in this project repository. Removing it
avoids a misleading second "README" and keeps the root `README.md` as the single
project entry point.

**Status banners.** Aspirational design docs receive a short status banner that
distinguishes target design from current implementation and links to
`docs/PRODUCTION_READINESS.md`. The design content itself is preserved unchanged.

**Vendor-name decoupling.** Workflow-visible short-form activity methods are
renamed to capability names (with `Mock`/`Real` role suffixes where a
test-double-vs-real split exists), removing vendor names from the
deterministic workflow layer and from registered Temporal activity names. This is
a semantics-preserving rename only:

- `GenerateMockVisualShots` → `GenerateVisualShotsMock`
- `GenerateSeedanceShots` → `GenerateVisualShotsReal`
- `GenerateElevenLabsVoiceover` → `GenerateVoiceover`
- `GenerateUploadPostPublishManifest` → `GeneratePublishManifest`
- `ValidateUploadPostPublishManifest` → `ValidatePublishManifest`
- `UploadPostDryRun` (activity) → `PublishDryRun`
- `UploadPostSchedulePublish` → `PublishSchedule`

The mock-vs-real split and every fail-closed behavior are preserved:
`GenerateVisualShotsReal` and `PublishSchedule` still refuse. The
provider-layer interface method `PublishingProvider.UploadPostDryRun` and the
`providers/uploadpost`, `providers/voice/omnivoice`, `providers/review/claude`
packages are intentionally **not** renamed: the provider layer is the deliberate
vendor boundary and is allowed to name vendors. A wiring test now guards that no
registered activity name encodes a commercial vendor.

## Consequences

- No behavioral change to gates, the publish-path invariant, immutability,
  self-approval, AI-disclosure, multi-verifier, or workflow semantics.
- Renaming the activity methods changes their registered Temporal activity names;
  all workflow call sites, the in-process runner, the test-environment
  registrations, and the worker registration are updated in the same change, and
  `make verify` (including the offline Temporal test environment) proves no name
  drift.
- The legal entity name in `LICENSE` is "Animus". If the exact registered legal
  entity differs, the owner must confirm and update it (tracked in the ledger,
  not in `LICENSE`).

## Alternatives considered

- **Keep `profile/` with an explanatory ADR (option b).** Rejected: it leaves
  inert org-template content in a project repo and a second README-like file.
  Removal is cleaner and the rationale is captured here.
- **Collapse mock/real activities into one dispatching activity.** Out of scope;
  it is a separate design decision that could weaken fail-closed guarantees and
  would need its own ADR.
- **Rename provider-layer vendor symbols too.** Rejected: the provider layer is
  the sanctioned vendor boundary; renaming there expands scope without benefit.
