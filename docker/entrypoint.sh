#!/usr/bin/env bash

set -o errtrace -o nounset -o pipefail -o errexit

echo "uname -a: $(uname -a)"

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

for entry in ./packer-*; do
    if [[ -x "$entry" ]]; then
        echo "Please remove $entry and try again."
        exit 1
    fi
done

export DONT_SETUP_QEMU=1

# ensure packer plugin/cache directories exist
mkdir -p "${PACKER_PLUGIN_PATH}"
mkdir -p "${PACKER_CACHE_DIR}"

echo running "${PACKER}" "${@}"

exec "${PACKER}" "${@}"
