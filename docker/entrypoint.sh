#!/usr/bin/env bash

set -o errtrace -o nounset -o pipefail -o errexit

bash /register --reset -p yes >/dev/null 2>&1

PACKER=/bin/packer

echo running "${PACKER}"

exec "${PACKER}" "${@}"
