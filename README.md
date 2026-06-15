# AnimusHQ organization defaults

This repository stores organization-level GitHub defaults for AnimusHQ.

The public organization profile is maintained in:

```text
profile/README.md
```

Default community and governance files in this repository apply to AnimusHQ repositories unless a repository provides its own more specific version.

## Organization scope

AnimusHQ builds secure access, control-plane, management-plane, validation, and observability components for device-connected and private infrastructure.

The organization does not present early repositories as production-ready security products unless they define:

- supported deployment scope;
- threat model;
- security policy;
- release process;
- CI gates;
- vulnerability reporting path;
- operational documentation.

## Repository baseline

Each public repository should define:

- `README.md` with scope, status, non-goals, and local development path;
- `LICENSE` or explicit licensing status;
- `SECURITY.md` or inherited security policy;
- `CONTRIBUTING.md` or inherited contribution rules;
- issue and pull request templates;
- CI for formatting, linting, and tests;
- architecture notes for non-trivial runtime systems.

## Contact

Security reports and collaboration inquiries: [rewanderer@proton.me](mailto:rewanderer@proton.me)
