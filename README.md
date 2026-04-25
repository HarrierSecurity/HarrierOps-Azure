# HarrierOps Azure

<p align="center">
  <img src="assets/ho-azure-logo.png" alt="HarrierOps Azure logo" height="280" />
</p>

HarrierOps Azure is an Azure reconnaissance CLI for offensive security professionals who want to see
how ordinary Azure control-plane features become persistence, evasion, resource hijacking, path
masking, and chained movement opportunities.

It helps you move past inventory and expose the uncomfortable part of cloud security: the same
automation, identity, telemetry, and routing features defenders rely on can also become the paths an
operator needs to understand first.

Try it in the release container:

```bash
docker run --rm ghcr.io/harriersecurity/ho-azure:v1.2.0 help
```

HarrierOps Azure helps you answer:

- Where could someone leave a durable way back in through Automation, App Services, Logic Apps,
  WebJobs, or other trusted Azure runtimes?
- Where could logging be turned down, rerouted, filtered, or quietly made less useful?
- Which existing Azure resources could be repurposed instead of creating something suspicious and
  new?
- Which trusted workflows, gateways, relays, or connectors could make activity look like it came
  from somewhere else?
- Which identities, permissions, and resource relationships make those moves possible from the
  current foothold?
- Where does the evidence stop because of permissions, Azure visibility, or source boundaries?

## Operator Focus

HarrierOps Azure is operator-forward rather than inventory-first. It pulls the interesting Azure
control paths to the top so you are not stuck sorting raw resources before you know what matters.

In cloud environments, the useful path is not always a classic persistence trick or a loud new
resource. It can be a runbook that already looks like maintenance, an app that already has a trusted
identity, a workflow that already talks to downstream services, a logging route that can be made
less useful, or a relay/gateway path that makes activity look like normal platform plumbing.

Flat commands collect the evidence. Grouped command families turn that evidence into paths:
privilege escalation, persistence, evasion, resource hijacking, path masking, and chained Azure
movement.

The goal is not to claim more than Azure proves. Output is shaped around truth boundaries: what the
current identity can defend, what is merely visible, and where reduced visibility should stop the
story instead of becoming a misleading empty result.

If you want both sides of the story, run the companion proof lab: use HarrierOps Azure to see the
operator path, then review the Azure-side logs the lab generates to understand what defenders can
and cannot see.

## Install

Option 1: run the release container:

```bash
docker run --rm ghcr.io/harriersecurity/ho-azure:v1.2.0 help
```

Replace `v1.2.0` with the latest release tag when a newer release is available.

Option 2: download the latest binary release for your platform:

[HarrierOps Azure releases](https://github.com/HarrierSecurity/HarrierOps-Azure/releases/latest)

Option 3: install with the HarrierOps Homebrew tap:

```bash
brew install harriersecurity/ho-azure/ho-azure
```

Option 4: build from source:

```bash
git clone https://github.com/HarrierSecurity/HarrierOps-Azure.git
cd HarrierOps-Azure
go build -o ho-azure ./cmd/azurefox
./ho-azure help
```

See [Getting Started](https://github.com/HarrierSecurity/HarrierOps-Azure/wiki/Getting-Started#1-install)
for install profile notes and development commands.

## Operator Workflow

Start with the identity you have, then work outward toward movement and consequence.

Typical flow:

```bash
ho-azure whoami
ho-azure permissions
ho-azure privesc
ho-azure persistence automation
ho-azure evasion dcr
```

- `whoami`: confirm the current foothold, token context, and subscription scope.
- `permissions`: identify where that identity already has meaningful control.
- `privesc`: surface direct abuse or escalation paths rooted in the current access.
- `persistence automation`: check whether trusted runbooks, schedules, webhooks, identities, or
  worker context can preserve or re-trigger access.
- `evasion dcr`: check whether Data Collection Rules can reshape collection, routing,
  destinations, associations, or transformations from the management plane.

## Operator Outcome

After one pass, you should know:

- which identity you are actually holding
- what that identity can control right now
- whether privilege escalation is already visible
- whether durable automation can preserve or re-trigger access
- whether telemetry collection can be quietly reshaped from visible management-plane control

HarrierOps Azure reduces noise by ranking consequence, not just returning Azure objects.

## Currently Supported Azure Commands

### Orchestration

| Grouped Command | Live Families |
| --- | --- |
| `chains`<br>Grouped path views that pull the strongest Azure pivot stories to the top. | `credential-path`, `deployment-path`, `escalation-path`, `compute-control` |
| `persistence`<br>Service-specific persistence walkthroughs focused on what the current identity can preserve, trigger, or reuse. | `app-service`, `automation`, `azure-ml`, `container-apps-jobs`, `functions`, `logic-apps`, `vm-extensions`, `webjobs` |
| `evasion`<br>Service-specific views of quiet Azure-native defender-truth disruption. | `appinsights`, `dcr`, `diagnostic-settings` |
| `resourcehijacking`<br>Service-specific takeover views for commandeering, redirecting, replacing, or repurposing trusted resources. | `api-mgmt`, `automation`, `logic-apps` |
| `pathmasking`<br>Service-specific relay, proxy, and workflow views for path ambiguity and attribution blur. | `api-mgmt`, `logic-apps`, `relay` |

### Flat Commands

| Section | Commands |
| --- | --- |
| `core` | `inventory` |
| `identity` | `whoami`, `rbac`, `principals`, `permissions`, `privesc`, `role-trusts`, `lighthouse`, `cross-tenant`, `auth-policies`, `managed-identities` |
| `config` | `arm-deployments`, `env-vars` |
| `secrets` | `keyvault`, `tokens-credentials` |
| `resource` | `automation`, `devops`, `acr`, `api-mgmt`, `appinsights`, `databases`, `dcr`, `diagnostic-settings`, `monitoring-sinks`, `resource-trusts` |
| `storage` | `storage` |
| `network` | `application-gateway`, `nics`, `dns`, `endpoints`, `network-effective`, `network-ports`, `relay` |
| `compute` | `workloads`, `app-services`, `functions`, `container-apps`, `container-apps-jobs`, `container-instances`, `aks`, `vms`, `vm-extensions`, `vmss`, `snapshots-disks` |

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

See the [Getting Started auth section](https://github.com/HarrierSecurity/HarrierOps-Azure/wiki/Getting-Started#2-authenticate)
for setup examples for Azure CLI users, service principals, environment credentials, managed
identities, and Azure DevOps organization targeting.

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

Help also points grouped follow-up toward `chains`, `persistence`, `evasion`,
`resourcehijacking`, and `pathmasking` where those presets exist.

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

## FAQ

### Is HarrierOps Azure read-only?

Yes. HarrierOps Azure is built as a read-only reconnaissance tool. It queries Azure and related
control-plane surfaces to show identity, permission, resource, and path relationships. It does not
create, update, delete, or execute Azure resources.

### What makes HarrierOps Azure different from normal inventory tools?

Most inventory tools tell you what exists. HarrierOps Azure helps you figure out what kind of Azure
foothold you are actually holding and what that access can lead to.

It starts with identity-level questions: who am I in this tenant, what subscription am I looking at,
what permissions do I really have, and which resources trust this identity?

From there, it connects related evidence into operator-ready paths instead of leaving you with a
pile of raw Azure objects. The goal is to get from "I can see some Azure stuff" to "this identity
can preserve access here, weaken visibility there, repurpose that trusted resource, or follow this
chain next."

### What happens when my identity has limited permissions?

Your visibility in Azure depends heavily on the active login you are running the tool under. When
Azure blocks part of the picture, the output should say that instead of making the result look
empty.

A limited identity might still expose useful paths, but some paths may stop early because the next
hop, target, workflow, or permission check is not visible from the current access. The goal is to
separate "nothing found" from "Azure would not let this identity see farther."

### Where do output artifacts go?

Use `--outdir` to choose a run directory. If `--outdir` is not provided, artifacts are written in
the current directory. For ad hoc work, a dedicated path such as `--outdir ./ho-azure-demo` keeps
JSON, table, CSV, and future output changes out of the repo root.

### What is artifact-backed session reuse?

When helper artifacts from the same active workspace match the current tenant, subscription,
principal, auth context, tool/schema version, command options, and freshness window, grouped
commands can reuse that data instead of asking Azure the same question again. Reused artifacts are
session reuse with provenance, not fresh Azure truth.

### How was AI used to create HarrierOps Azure?

AI assisted with rapid prototyping, code generation, documentation drafts, and review passes during
development. HarrierOps Azure is not a one-shot vibe-coded project. The shipped tool is shaped by
planning notes, command contracts, system design decisions, deterministic fixtures, unit tests,
golden output checks, live-lab follow-up, release smoke tests, and repeated human review.

Development also included sustained review of what makes a recon tool useful in practice:
operator workflow, OPSEC, performance, output truth boundaries, reduced-visibility handling,
artifact provenance, packaging, and whether a command is actually worth shipping as a first-class
surface instead of just being an interesting idea.

The goal is for HarrierOps Azure to stand on its own as a serious operator tool: useful output,
clear truth boundaries, reproducible tests, and command behavior that can survive review instead of
just looking impressive in a demo.

## License

HarrierOps Azure is licensed under the MIT License. See [LICENSE](LICENSE).
