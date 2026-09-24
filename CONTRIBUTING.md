# Contributing

Awesome Reliability Tools is maintained through public GitHub pull requests. The structured dataset in `data/tools.yaml` is the only source of truth for the README and website.

## Before you contribute

Read [METHODOLOGY.md](METHODOLOGY.md) for the principles that guide this list.

The short version:

- No paid placement, rankings, or "best" labels.
- Factual, neutral descriptions only.
- Every feature claim should be backed by a primary source.
- Unknown / unverified is not the same as unsupported.
- Maintainer-owned products follow the exact same rules as any other entry.

## Adding or updating a tool

1. Fork the repository and create a branch.
2. Edit `data/tools.yaml`. Add a new entry or update an existing one.
3. Make sure the entry uses only feature IDs defined in `data/features.yaml`. If you need a new capability, add it to `data/features.yaml` in the same PR and explain why it is needed.
4. Provide source URLs for every feature you mark as `supported` or `unsupported`. Use the vendor's or project's own documentation, not third-party summaries.
5. Set `verified_at` to the date you confirmed the claim (YYYY-MM-DD).
6. Run the validator and generator locally:

   ```bash
   make validate
   make generate
   ```

7. Run the tests:

   ```bash
   make test
   ```

8. Commit both the data change and the generated `README.md`.
9. Open a pull request with a clear description of the change and the sources you used.

## Favicons

You do not need to add favicons. After a tool change is merged, the `Favicons` workflow fetches a small icon from the tool homepage and proposes it in an automated pull request, and a monthly run refreshes existing icons. Manually supplied favicon changes are usually unnecessary. Local runs of `make generate` never fetch favicons, so builds stay deterministic and offline.

## What belongs in the main directory

The project covers the service reliability lifecycle:

- **Detect**: uptime / availability monitoring (HTTP, HTTPS, TCP, DNS, ICMP, heartbeat, SSL, multi-region, API checks).
- **Respond**: incident management, automatic incident creation, on-call, escalation policies, maintenance windows, postmortems.
- **Communicate**: public and private status pages, components, custom domains, subscribers, RSS/Atom, uptime history.

Avoid adding generic observability, logging, tracing, or APM products unless they directly satisfy one of the areas above.

## Evidence quality

- Preferred: official documentation, public API docs, or the open-source repository.
- Avoid: blog posts, comparison sites, vendor marketing copy without specifics, or claims you have not verified.
- If you cannot find evidence, use `value: unknown` and explain in `notes`.

## Style

- Keep descriptions concise and factual (one or two sentences).
- Use the existing feature names and values.
- Do not add marketing superlatives such as "best", "leading", or "top".

## Governance

Maintainers will review PRs for neutrality, evidence, and schema correctness. We may ask for clearer sources or reject entries that fall outside the scope. Disagreements are resolved by referencing [METHODOLOGY.md](METHODOLOGY.md).

## Lifecycle status

Tools are `active` by default. Use `status: deprecated` or `status: archived` when a project is no longer being developed, and keep the entry so its history is preserved. Add the end-of-life date or replacement product to `notes` when known.

## License

The Go tooling is licensed under the MIT License (see `LICENSE`). The dataset and documentation are released under CC0 1.0 (see `LICENSE-CC0`). By contributing, you agree that your contributions are released under the corresponding license for the files you change.
