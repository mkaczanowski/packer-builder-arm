#!/usr/bin/env bash

set -o errtrace -o nounset -o pipefail -o errexit

echo "uname -a: " $(uname -a)

/usr/bin/binfmt --install all

PACKER=/bin/packer

declare -a EXTRA_SYSTEM_PACKAGES=()
for arg do
    shift
    if [[ "${arg}" == -extra-system-packages=* ]]; then
        IFS=',' read -r -a EXTRA_SYSTEM_PACKAGES <<< "${arg//-extra-system-packages=}"
        continue
    fi
    set -- "$@" "${arg}"
done

if [ "${#EXTRA_SYSTEM_PACKAGES[@]}" -gt 0 ]; then
    echo "Installing extra system packages: ${EXTRA_SYSTEM_PACKAGES[*]}"
    apt-get update
    apt-get install -y --no-install-recommends "${EXTRA_SYSTEM_PACKAGES[@]}"
fi

export DONT_SETUP_QEMU=1

echo running "${PACKER}" "${@}"

exec "${PACKER}" "${@}"
