# Contributing to Animus News

Animus News is **proprietary software**. All rights are retained exclusively by
Animus. See [`LICENSE`](LICENSE). The public visibility of this repository does
not grant any license and does not make this an open-source project.

## Contribution policy

External contributions are **not accepted by default**. Because the project is
proprietary, code, documentation, and other materials in this repository may not
be copied, modified, or redistributed without the prior express written
permission of Animus.

If Animus expressly invites a contribution, it is accepted **only** under a prior
written agreement assigning all rights, title, and interest in the contribution
to Animus. Without such an agreement, please do not open pull requests.

This policy applies to this repository only. Separately published Animus
open-source community projects are governed by their own contribution rules and
license terms.

## Reporting issues

You may still report problems without contributing code:

- **Security vulnerabilities:** report privately per [`SECURITY.md`](SECURITY.md).
  Do not open public issues for vulnerabilities.
- **Bugs and documentation defects:** include the affected commit, environment,
  expected vs. actual behavior, and minimal reproduction steps. Do not include
  secrets, tokens, private keys, customer data, or internal logs.

## Working agreements (for authorized contributors)

When a contribution is authorized in writing, follow the repository rules in
[`AGENTS.md`](AGENTS.md) and [`CLAUDE.md`](CLAUDE.md):

- keep changes scoped to an explicit task pack;
- do not weaken gates, the publish-path invariant, or the safety model;
- include tests or reproducible validation, and keep `make verify` green;
- never add secrets, live provider calls, spend, or public publishing;
- keep documented status honest (Implemented / Partial / Planned).
