# HarrierOps Azure

Operator-first Azure recon in Go, with the `azurefox` CLI surface preserved for parity with AzureFox.

HarrierOps Azure is built to answer the practical questions that matter after you already have Azure
access:

- who am I
- what resource surface is visible
- which identities, trusts, platforms, and control paths matter first
- where the strongest follow-up path is likely to be

## Current Shape

The binary name is `azurefox`.

This repo currently supports the flat AzureFox-style command surface in Go, including core,
identity, compute, network, resource, config, and trust-oriented commands. Use `azurefox help` to
see the current implemented command list.

`chains` family work is planned separately from the flat-command surface.

## Install

Build directly from source:

```bash
go build -o azurefox ./cmd/azurefox
```

## Provider Modes

By default, the CLI uses the static provider:

```bash
./azurefox whoami --output table
```

To use live Azure collection instead, set `AZUREFOX_PROVIDER=azure`:

```bash
AZUREFOX_PROVIDER=azure ./azurefox inventory --output json
```

The live Azure path uses the repo's Azure provider layer directly. It prefers Azure CLI-backed
authentication and can also use supported Azure environment credential settings when present.

## Operator Workflow

Start by confirming identity and visible scope, then move into the surfaces that answer what that
access can actually reach:

```bash
./azurefox whoami
./azurefox inventory
./azurefox permissions
./azurefox role-trusts
./azurefox privesc
```

Typical flow:

- `whoami`: confirm the current tenant, subscription, and principal context
- `inventory`: determine the visible Azure surface quickly
- `permissions`, `rbac`, `managed-identities`, `role-trusts`: understand what identity-driven
  access paths matter
- `privesc`, `cross-tenant`, `resource-trusts`: surface the strongest follow-up paths

## DevOps Note

The `devops` command accepts `--devops-organization` and also honors `AZUREFOX_DEVOPS_ORG` for the
organization name when live Azure DevOps collection is needed.

## Role Trusts Note

`role-trusts` defaults to `fast` mode. The command also supports `full`, plus the deprecated
comparison modes `fast-old` and `full-old` for parity testing.

## Output Modes

- `--output table` (default)
- `--output json`
- `--output csv`

When `--outdir` is set, the CLI writes artifacts under:

- `loot/<command>.json`
- `json/<command>.json`
- `table/<command>.txt`
- `csv/<command>.csv`

## CLI Shape

Commands come first, then flags:

```bash
azurefox <command> [flags]
```

Examples:

```bash
./azurefox help
./azurefox whoami --output json
AZUREFOX_PROVIDER=azure ./azurefox inventory --subscription <subscription-id> --output table
AZUREFOX_PROVIDER=azure ./azurefox devops --devops-organization <org> --output json
AZUREFOX_PROVIDER=azure ./azurefox role-trusts --mode fast --output json
```

## Development

```bash
gofmt -w ./cmd ./internal
go test ./...
bash scripts/setup_local_guardrails.sh
```
