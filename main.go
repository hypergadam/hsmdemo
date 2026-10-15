// Command hsmdemo shows how to create an ECDSA P-384 key inside a PKCS#11
// token (for example SoftHSM), sign data with it, and verify the signature.
//
// The private key is created inside the HSM and can never be exported.
//
// Environment variables:
//
//	HSM_PIN     user PIN of the token (required)
//	PKCS11_LIB  path to the PKCS#11 library (default: Fedora SoftHSM path)
package main

import (
	"crypto/ecdsa"
	"crypto/sha512"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/miekg/pkcs11"
)

const (
	defaultLibPath = "/usr/lib64/pkcs11/libsofthsm2.so" // Fedora path
	tokenLabel     = "ca-token"
	keyLabel       = "ca-signing-key"

	// P-384 numbers are 48 bytes long.
	coordSize = 48
)

func main() {
	if err := run(); err != nil {
		slog.Error("hsm demo failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	libPath := os.Getenv("PKCS11_LIB_PATH")
	if libPath == "" {
		libPath = defaultLibPath
	}

	pin := os.Getenv("HSM_PIN")
	if pin == "" {
		return errors.New("HSM_PIN is not set")
	}

	// 1. Load the PKCS#11 library and start it.
	p := pkcs11.New(libPath)
	if p == nil {
		return fmt.Errorf("cannot load PKCS#11 library %q", libPath)
	}
	defer p.Destroy()

	if err := p.Initialize(); err != nil {
		return fmt.Errorf("initialize module: %w", err)
	}
	defer p.Finalize()

	// 2. Find the slot by token label (slot numbers can change).
	slot, err := findSlot(p, tokenLabel)
	if err != nil {
		return err
	}

	// 3. Open a session and log in.
	session, err := p.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		return fmt.Errorf("open session: %w", err)
	}
	defer p.CloseSession(session)

	if err := p.Login(session, pkcs11.CKU_USER, pin); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	defer p.Logout(session)

	// 4. Find the key. If it does not exist, create it inside the HSM.
	privKey, err := findObject(p, session, pkcs11.CKO_PRIVATE_KEY, keyLabel)
	if err != nil {
		return err
	}
	pubObject, err := findObject(p, session, pkcs11.CKO_PUBLIC_KEY, keyLabel)
	if err != nil {
		return err
	}

	if privKey == 0 {
		pubObject, privKey, err = generateKeyPair(p, session)
		if err != nil {
			return err
		}
		slog.Info("new key pair created in HSM")
	}
	if pubObject == 0 {
		return errors.New("private key found, but public key is missing")
	}

	// 5. Hash the data, then ask the HSM to sign the hash.
	// SHA-384 is the natural partner of P-384. SHA-512 also works.
	digest := sha512.Sum384([]byte("hello from my CA"))

	err = p.SignInit(session,
		[]*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_ECDSA, nil)}, privKey)
	if err != nil {
		return fmt.Errorf("sign init: %w", err)
	}
	sig, err := p.Sign(session, digest[:]) // 96 bytes: r (48) + s (48)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	if len(sig) != 2*coordSize {
		return fmt.Errorf("unexpected signature length: %d", len(sig))
	}

	// 6. Read the public key from the HSM.
	pub, err := readPublicKey(p, session, pubObject)
	if err != nil {
		return err
	}

	// 7. Verify the signature.
	r := new(big.Int).SetBytes(sig[:coordSize])
	s := new(big.Int).SetBytes(sig[coordSize:])
	ok := ecdsa.Verify(pub, digest[:], r, s)

	slog.Info("signature checked", slog.Bool("valid", ok))
	if !ok {
		return errors.New("signature is NOT valid")
	}

	return nil
}
