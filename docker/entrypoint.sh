#!/usr/bin/env bash

set -o errtrace -o nounset -o pipefail -o errexit

/usr/sbin/update-binfmts --enable qemu-arm >/dev/null 2>&1

PACKER=/bin/packer

echo running "${PACKER}"

exec "${PACKER}" "${@}"
