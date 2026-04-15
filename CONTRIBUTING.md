# Contributing

## Local Setup

HarrierOps Azure is a Go-first repo with fixture-backed contract and output tests right now.

```bash
bash scripts/setup_local_guardrails.sh
```

If you want local secret scanning before push, install `gitleaks` separately. The pre-push hook
will run it when available and will otherwise leave that check to CI.

## Default Validation

```bash
gofmt -w ./cmd ./internal
go test ./...
```

## Test Shape

- CLI smoke and artifact tests: command execution against static provider output
- Contract tests: command metadata and family invariants
- Provider tests: embedded fixture-backed provider and command-output stability

## Documentation Boundary

- Keep operator-facing documentation in the repo.
- Keep package, build, install, schema, and release metadata in the repo when the repo needs it to
  build, validate, publish, or package correctly.
- Move maintainer-only planning notes and reference material into the workspace-level reference area
  instead of keeping them under the repo tree.
- Do not commit repo-local planning notes when the reference area outside the repo can carry them.

## Semantics And Contracts

- Keep command boundaries stable.
- Keep JSON output deterministic.
- Update golden outputs in the same change when a command contract moves.
- Keep operator wording aligned with the repo's evidence-boundary and truthfulness standards.

## Lightweight Guardrails

- Create a short-lived branch per change, such as `feat/...`, `fix/...`, or `docs/...`.
- Open a PR into `main` even when working solo.
- Keep PRs small and single-purpose.
- Merge only after CI is green.
- If command output contracts change, update the relevant golden fixtures in the same PR.
- Local pre-push hook blocks `codex` branch names, blocks direct pushes to `main`, checks
  formatting, runs tests, and runs `gitleaks` when installed locally.
- CI blocks Codex-branded PR titles and runs `gitleaks` plus the Go validation suite.
- Temporary bypass for emergency push: `HARRIEROPS_AZURE_ALLOW_MAIN_PUSH=1 git push`
