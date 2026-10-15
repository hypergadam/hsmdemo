#!/bin/sh
# Create the SoftHSM token if it is not there yet, then run hsmdemo.
#
# This is the container equivalent of `just token-ensure`: the image ships no
# token, so a fresh container (or a fresh volume) has to initialise one before
# the program can find it.
set -eu

# "ca-token" must match the tokenLabel constant in main.go.
token_label='ca-token'

: "${HSM_PIN:?HSM_PIN is required (user PIN of the token)}"

# Fall back to locating the provider if PKCS11_LIB_PATH is unset or wrong, so the
# image keeps working if the distro moves libsofthsm2.so.
if [ -z "${PKCS11_LIB_PATH:-}" ] || [ ! -f "${PKCS11_LIB_PATH}" ]; then
    # busybox find has no -quit, so take the first hit with head.
    PKCS11_LIB_PATH="$(find /usr/lib -name libsofthsm2.so 2>/dev/null | head -n 1)"
    if [ -z "${PKCS11_LIB_PATH}" ]; then
        echo 'docker-entrypoint: no libsofthsm2.so found in /usr/lib' >&2
        exit 1
    fi
    export PKCS11_LIB_PATH
fi

if ! softhsm2-util --show-slots 2>/dev/null | grep -q "${token_label}"; then
    : "${HSM_SO_PIN:?HSM_SO_PIN is required to create the token on first start}"
    echo "docker-entrypoint: initialising token \"${token_label}\""
    softhsm2-util --init-token --free \
        --label "${token_label}" \
        --so-pin "${HSM_SO_PIN}" \
        --pin "${HSM_PIN}"
else
    echo "docker-entrypoint: token \"${token_label}\" already present"
fi

exec hsmdemo "$@"
