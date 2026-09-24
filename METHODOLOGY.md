# Methodology

This document explains how Awesome Reliability Tools selects, verifies, and presents entries.

## Scope

We catalogue tools that participate in the service reliability lifecycle:

```
Detect → Respond → Communicate
```

### Detect

Tools that watch services for problems:

- HTTP/HTTPS monitoring
- TCP, DNS, and ICMP/ping checks
- API monitoring and assertions
- SSL certificate monitoring
- Heartbeat / push monitoring
- Multi-region or globally distributed checks

### Respond

Tools that coordinate the reaction to problems:

- Incident management workflows
- Automatic incident creation
- On-call scheduling and rotation
- Escalation policies
- Maintenance windows
- Postmortem support

### Communicate

Tools that share service health with users and stakeholders:

- Public status pages
- Private status pages
- Component-level status
- Custom domains
- Subscriber notifications
- RSS/Atom feeds
- Uptime history display

## Neutrality

- No rankings, leaderboards, "best" labels, scores, or paid placement.
- Every tool is rendered using the same schema and rules.
- Maintainer-owned products receive no special treatment.
- Descriptions are factual and avoid marketing language.

### Maintainer disclosure

This project is maintained by the founder of [StatusPage.me](https://statuspage.me/). StatusPage.me is included and documented using the same criteria, evidence requirements, sorting, and presentation as every other tool. It receives no featured placement or preferential treatment.

## Evidence and verification

The dataset records the state of each capability as one of:

- `supported`
- `unsupported`
- `unknown / unverified`
- `not_applicable`

A claim of `supported` or `unsupported` must include a source URL and a `verified_at` date. Sources should be primary: official documentation, API references, or the project's own repository. We do not treat "not found in documentation" as automatically `unsupported`; if evidence is missing, the state should be `unknown / unverified`.

Verification dates make freshness visible. Older dates are not automatically invalid, but they signal that the claim should be re-checked.

## Schema as source of truth

`data/tools.yaml` and `data/features.yaml` are the canonical sources. The following are generated and must not be hand-edited:

- `README.md`
- `dist/data/catalog.json`
- Static site files under `dist/`

## Lifecycle

Every tool carries an optional lifecycle status:

- `active` (default)
- `deprecated` — still usable, but end of life is announced
- `archived` — no longer developed

Discontinued tools are retained and labelled rather than silently deleted, so their history and evidence stay available. A future phase may group them into a dedicated historical section.

## Provenance

Initial discovery was informed by these existing awesome lists, used only as reference sources. Descriptions and data were independently rewritten and verified against primary sources where possible.

- [AlertSleep/awesome-uptime-monitoring](https://github.com/AlertSleep/awesome-uptime-monitoring)
- [KHiis/awesome-incident-management](https://github.com/KHiis/awesome-incident-management)
- [ivbeg/awesome-status-pages](https://github.com/ivbeg/awesome-status-pages)

## Updates

Corrections and additions happen through public GitHub pull requests. Vendors, users, and maintainers all follow the same process described in [CONTRIBUTING.md](CONTRIBUTING.md).
