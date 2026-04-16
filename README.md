# HarrierOps Azure

Find attack paths, pivot opportunities, and movement across Azure before you drown in inventory.

Most Azure tools tell you what exists.
HarrierOps Azure tells you how an identity can move between those resources.
Most Azure tools dump permissions.
HarrierOps Azure highlights which relationships, pivots, and escalation paths matter first.

The shipped CLI binary is `ho-azure`.

## Why This Matters

You have:

- a compromised user
- service principal access
- a managed identity foothold
- partial subscription visibility

You need to answer quickly:

- What identity am I actually holding?
- What can it control right now?
- Where can it pivot next?
- Which path is most likely to become privilege escalation or broader Azure control?

HarrierOps Azure is built for that workflow.

## Why This Is Different

- Attack-path thinking, not inventory-first reporting
- Pivot-first workflow, not isolated command output
- Identity and permission relationships, not just raw role listings
- Operator guidance that points to the next path worth investigating
- Broader than a foothold check: useful for movement, consequence, and follow-on access across Azure

## Core Capabilities

- Show the active Azure identity, token context, and scope you are operating from
- Surface high-impact RBAC and permission relationships that change what the current identity can do
- Map identity trust, service principal ownership, federated credentials, and cross-tenant edges
- Highlight pivot paths through workloads, managed identities, deployment systems, and secret-bearing configuration
- Expose escalation opportunities and likely next steps instead of leaving you to sort raw Azure data

## Install

Build from source:

```bash
go build -o ho-azure ./cmd/azurefox
```

If you prefer to run without creating a local binary first:

```bash
go run ./cmd/azurefox help
```

## Operator Workflow

Start with the identity you have, then work outward toward movement and consequence.

Typical flow:

- `whoami`: confirm the current foothold, token context, and subscription scope
- `permissions`: identify where that identity already has meaningful control
- `privesc`: surface direct abuse or escalation paths rooted in the current access
- `role-trusts` and `cross-tenant`: find identity-control transforms and tenant boundary pivots
- `tokens-credentials` and `chains`: follow token, secret, and deployment clues toward the next usable path

## Operator Outcome

After one pass, you should know:

- which identity matters
- what access is real versus merely visible
- where the best pivot opportunities are
- which attack path deserves follow-up first

HarrierOps Azure reduces noise by ranking consequence, not just returning Azure objects.

## Use Cases

- Triage a compromised user, service principal, or managed identity and determine what Azure control it enables
- Assess whether a service principal or application relationship creates a pivot or escalation path
- Work outward from subscription or tenant visibility to identify cross-resource and cross-tenant movement

## Run It

Start with the current Azure identity and the strongest visible control paths:

```bash
ho-azure whoami
ho-azure permissions
```

## Currently Supported Azure Commands

### Orchestration

| Grouped Command | Live Families |
| --- | --- |
| `chains`<br>Grouped path views that pull the strongest Azure pivot stories to the top. | `credential-path`<br>Turns exposed secret and token clues into the downstream target most likely to widen access.<br><br>`deployment-path`<br>Surfaces the build, pipeline, and automation paths most likely to let an attacker change Azure next.<br><br>`escalation-path`<br>Highlights the clearest visible route from the current foothold to stronger Azure control.<br><br>`compute-control`<br>Finds workloads that can already mint identity-backed access and pivot into broader control. |

### Flat Commands

| Section | Commands |
| --- | --- |
| `core` | `inventory` |
| `identity` | `whoami`, `rbac`, `principals`, `permissions`, `privesc`, `role-trusts`, `lighthouse`, `cross-tenant`, `auth-policies`, `managed-identities` |
| `config` | `arm-deployments`, `env-vars` |
| `secrets` | `keyvault`, `tokens-credentials` |
| `resource` | `automation`, `devops`, `acr`, `api-mgmt`, `databases`, `resource-trusts` |
| `storage` | `storage` |
| `network` | `application-gateway`, `nics`, `dns`, `endpoints`, `network-effective`, `network-ports` |
| `compute` | `workloads`, `app-services`, `functions`, `container-apps`, `container-instances`, `aks`, `vms`, `vmss`, `snapshots-disks` |

## Need A Test Lab?

Use the companion HarrierOps Azure lab repo for live validation when you want backend-honest behavior.
The static provider in this repo is for deterministic local inspection and tests, not a substitute for
a real Azure-backed lab.

## CLI Invocation

Shared flags like `--tenant`, `--subscription`, `--output`, `--outdir`, `--debug`, and
`--devops-organization` work before or after the command.

These forms are equivalent:

```bash
ho-azure dns --output json --outdir ./ho-azure-demo
ho-azure --output json --outdir ./ho-azure-demo dns
```

Use `ho-azure <command> --help` or `ho-azure help <command>` for command-specific help.

## Install Profiles

HarrierOps Azure builds the live Azure runtime path by default, so a normal source build is ready
for real Azure command execution.

For a local binary:

```bash
go build -o ho-azure ./cmd/azurefox
```

For direct execution from a checkout:

```bash
go run ./cmd/azurefox whoami
```

For local development:

```bash
go test ./...
```

HarrierOps Azure is intended to work on macOS, Linux, and Windows. The command examples below use
portable relative paths like `./ho-azure-demo`; shell syntax mainly differs for environment-variable
export and binary invocation.

Live operator guidance is built into `ho-azure help` and `ho-azure help <command>`.

- `go build -o ho-azure ./cmd/azurefox`
  builds the normal operator binary from a local checkout
- `go run ./cmd/azurefox ...`
  runs the same live Azure command profile directly from source
- `go test ./...`
  runs the contributor validation baseline for the Go repo

## Auth Precedence

1. Azure CLI credential
2. Environment credential

### Supported auth matrix

| Path | How it starts | Current support | Metadata `auth_mode` |
| --- | --- | --- | --- |
| Interactive user via Azure CLI | `az login` | supported | `azure_cli` |
| Service principal via Azure CLI | `az login --service-principal ...` | supported through Azure CLI | `azure_cli` |
| Managed identity via Azure CLI | `az login --identity` | supported through Azure CLI | `azure_cli` |
| Service principal via environment client secret | `AZURE_TENANT_ID` + `AZURE_CLIENT_ID` + `AZURE_CLIENT_SECRET` | supported | `environment` |
| Service principal via environment certificate | `AZURE_TENANT_ID` + `AZURE_CLIENT_ID` + `AZURE_CLIENT_CERTIFICATE_PATH` | supported | `environment` |
| Environment fallback after Azure CLI failure | automatic fallback when CLI auth is unavailable but environment auth succeeds | supported | `environment_fallback` |

HarrierOps Azure does not launch its own browser or managed-identity login flow. It relies on Azure Identity:

- `AzureCliCredential` for the active Azure CLI sign-in state
- `EnvironmentCredential` for supported service principal environment variables

### Interactive user via Azure CLI

If you want web-based authentication, run `az login` first outside HarrierOps Azure, then run
`ho-azure`.

Azure CLI example:

```bash
az login
az account set --subscription <subscription-id>
ho-azure inventory --subscription <subscription-id>
```

### Service principal via Azure CLI

This is useful for headless automation that still wants Azure CLI to hold the active login state.

With a client secret:

```bash
az login --service-principal \
  --username <client-id> \
  --password <client-secret> \
  --tenant <tenant-id>
az account set --subscription <subscription-id>
ho-azure whoami --subscription <subscription-id>
```

With a certificate:

```bash
az login --service-principal \
  --username <client-id> \
  --certificate /path/to/certificate.pem \
  --tenant <tenant-id>
az account set --subscription <subscription-id>
ho-azure whoami --subscription <subscription-id>
```

### Service principal via environment client secret

If you do not want to use Azure CLI login state, set service principal environment variables and
pass CLI flags for tenant or subscription targeting.

Environment client-secret example:

```bash
# macOS/Linux
export AZURE_TENANT_ID=<tenant-id>
export AZURE_CLIENT_ID=<client-id>
export AZURE_CLIENT_SECRET=<client-secret>
export AZUREFOX_DEVOPS_ORG=<org-name> # only needed for the devops command
ho-azure whoami --tenant <tenant-id> --subscription <subscription-id>
```

```powershell
# Windows PowerShell
$env:AZURE_TENANT_ID="<tenant-id>"
$env:AZURE_CLIENT_ID="<client-id>"
$env:AZURE_CLIENT_SECRET="<client-secret>"
$env:AZUREFOX_DEVOPS_ORG="<org-name>" # only needed for the devops command
ho-azure whoami --tenant <tenant-id> --subscription <subscription-id>
```

### Service principal via environment certificate

```bash
# macOS/Linux
export AZURE_TENANT_ID=<tenant-id>
export AZURE_CLIENT_ID=<client-id>
export AZURE_CLIENT_CERTIFICATE_PATH=/path/to/certificate.pem
export AZURE_CLIENT_CERTIFICATE_PASSWORD=<optional-password>
ho-azure whoami --tenant <tenant-id> --subscription <subscription-id>
```

```powershell
# Windows PowerShell
$env:AZURE_TENANT_ID="<tenant-id>"
$env:AZURE_CLIENT_ID="<client-id>"
$env:AZURE_CLIENT_CERTIFICATE_PATH="C:\\path\\to\\certificate.pem"
$env:AZURE_CLIENT_CERTIFICATE_PASSWORD="<optional-password>"
ho-azure whoami --tenant <tenant-id> --subscription <subscription-id>
```

### Azure-hosted managed identity via Azure CLI

This works when you are running on an Azure resource that already has a managed identity attached.

```bash
az login --identity
az account set --subscription <subscription-id>
ho-azure whoami --subscription <subscription-id>
```

For a user-assigned managed identity:

```bash
az login --identity --client-id <user-assigned-managed-identity-client-id>
az account set --subscription <subscription-id>
ho-azure whoami --subscription <subscription-id>
```

`AZUREFOX_DEVOPS_ORG` is only needed when running the `devops` command. The identity used for
`devops` still needs access to the Azure DevOps organization, not just ARM access to the tenant or
subscription.

## Output Modes

- `--output table` (default)
- `--output json`
- `--output csv`

All commands write artifacts under `<outdir>/`:

- `loot/<command>.json`
- `json/<command>.json`
- `table/<command>.txt`
- `csv/<command>.csv`

Artifact intent:

- `json/` is the full structured command record
- `loot/` is the smaller high-value handoff, focused on the top-ranked targets for quick operator
  follow-up and later chain-oriented workflows
- `table/` and `csv/` are convenience views rendered from the same underlying command result

## Sections And Chains

HarrierOps Azure keeps flat standalone commands and also supports grouped execution through `chains`.

For narrower current work:

- run the flat commands directly when you already know the lane you want
- use `chains` when you want a higher-value grouped answer instead of every source command on its own

Current section mappings:

- `identity`: `whoami`, `rbac`, `principals`, `permissions`, `privesc`, `role-trusts`, `lighthouse`, `cross-tenant`, `auth-policies`, `managed-identities`
- `config`: `arm-deployments`, `env-vars`
- `secrets`: `keyvault`, `tokens-credentials`
- `resource`: `automation`, `devops`, `acr`, `api-mgmt`, `databases`, `resource-trusts`
- `storage`: `storage`
- `network`: `application-gateway`, `nics`, `dns`, `endpoints`, `network-effective`, `network-ports`
- `compute`: `workloads`, `app-services`, `functions`, `container-apps`, `container-instances`, `aks`, `vms`, `vmss`, `snapshots-disks`
- `core`: `inventory`
- `orchestration`: `chains`

Current `chains` families:

- `credential-path`
- `deployment-path`
- `escalation-path`
- `compute-control`

## Help

HarrierOps Azure supports generic and scoped help:

```bash
ho-azure help
ho-azure help identity
ho-azure help permissions
ho-azure dns --help
ho-azure -h identity
ho-azure -h permissions
```

Command help includes ATT&CK cloud leads as investigation prompts, not proof that a technique
occurred.

Help also points grouped follow-up toward `chains` where those presets exist.

For ad hoc demos or local testing, use a dedicated path like `--outdir ./ho-azure-demo` so
artifacts do not pile up in the repo root.

## Static Provider Mode

Set `AZUREFOX_PROVIDER=static` to run against the deterministic static provider rather than live
Azure APIs.

```bash
# macOS/Linux
AZUREFOX_PROVIDER=static ho-azure rbac --output json
```

```powershell
# Windows PowerShell
$env:AZUREFOX_PROVIDER="static"
ho-azure rbac --output json
```

If `AZUREFOX_PROVIDER` is unset, HarrierOps Azure defaults to live Azure collection.

## Development

```bash
gofmt -w ./cmd ./internal
go test ./...
bash scripts/setup_local_guardrails.sh
```

CI should cover the deterministic command surfaces before release-gated changes move forward.

## Attribution

HarrierOps Azure builds on the AzureFox porting work and is inspired by [CloudFox](https://github.com/BishopFox/cloudfox), created by Bishop Fox.

## License

HarrierOps Azure is licensed under the MIT License. See [LICENSE](LICENSE).
