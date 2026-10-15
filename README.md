# hsmdemo

Small Go program that uses a PKCS#11 HSM (SoftHSM for testing) to:

1. create an ECDSA **P-384** key pair inside the HSM,
2. sign a **SHA-384** hash with the private key,
3. verify the signature with the public key.

The private key is created with `CKA_SENSITIVE=true` and
`CKA_EXTRACTABLE=false`, so it can never leave the HSM.

## Settings

| Variable     | Required | Meaning                                              |
|--------------|----------|------------------------------------------------------|
| `HSM_PIN`    | yes      | User PIN of the token                                |
| `HSM_SO_PIN` | Docker   | Security Officer PIN, used only to create the token  |
| `PKCS11_LIB` | no       | Path to the PKCS#11 library (default: Fedora SoftHSM) |

## Run with Docker or Podman

```bash
go mod init hsmdemo
go mod tidy            # creates go.sum

podman build -t hsmdemo .
podman run --rm -e HSM_PIN=1234 -e HSM_SO_PIN=5678 hsmdemo
```

(`docker` works the same way.)

To keep the token between runs, add a volume:

```bash
podman run --rm -v hsm-tokens:/home/app/tokens:Z \
  -e HSM_PIN=1234 -e HSM_SO_PIN=5678 hsmdemo
```

Expected output:

```
msg="new key pair created in HSM"
msg="signature checked" valid=true
```

## Run without Docker (Fedora)

```bash
sudo dnf install -y golang gcc libtool-ltdl-devel softhsm
softhsm2-util --init-token --free --label ca-token --so-pin 5678 --pin 1234
HSM_PIN=1234 go run .
```

If the library path is different, find it with:

```bash
rpm -ql softhsm | grep libsofthsm2.so
```

and set `PKCS11_LIB`.

## Security notes

- The PINs in the examples are for testing only. In production, use a secret
  manager (Podman/Docker secrets, Vault, etc.), not command-line values.
- SoftHSM is software. It is **not** FIPS 140-3 validated. For production,
  use a validated hardware HSM and set `PKCS11_LIB` to the vendor library.
- Raw ECDSA from PKCS#11 is `r || s`. X.509 needs ASN.1 DER format, so convert
  it before putting the signature into a certificate.