#!/usr/bin/env sh
set -eu

binary="${1:?usage: release_smoke_unix.sh /path/to/ho-azure}"

"$binary" help >/dev/null
AZUREFOX_PROVIDER=static "$binary" whoami --output json >/dev/null
AZUREFOX_PROVIDER=static "$binary" chains credential-path --output json >/dev/null
AZUREFOX_PROVIDER=static "$binary" persistence automation --output json >/dev/null
AZUREFOX_PROVIDER=static "$binary" evasion dcr --output json >/dev/null
AZUREFOX_PROVIDER=static "$binary" resourcehijacking api-mgmt --output json >/dev/null
AZUREFOX_PROVIDER=static "$binary" pathmasking relay --output json >/dev/null
