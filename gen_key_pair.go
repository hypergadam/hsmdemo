package main

import (
	"fmt"

	"github.com/miekg/pkcs11"
)

// generateKeyPair creates a P-384 key pair inside the HSM.
// The private key is sensitive and not extractable.
func generateKeyPair(p *pkcs11.Ctx, s pkcs11.SessionHandle) (pub, priv pkcs11.ObjectHandle, err error) {
	// DER bytes of the curve OID for P-384 (secp384r1).
	ecParams := []byte{0x06, 0x05, 0x2B, 0x81, 0x04, 0x00, 0x22}

	pubTmpl := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_EC),
		pkcs11.NewAttribute(pkcs11.CKA_EC_PARAMS, ecParams),
		pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, keyLabel),
	}
	privTmpl := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_PRIVATE, true),
		pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),    // cannot be read out
		pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false), // cannot be exported
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, keyLabel),
	}

	pub, priv, err = p.GenerateKeyPair(s,
		[]*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_EC_KEY_PAIR_GEN, nil)},
		pubTmpl, privTmpl)
	if err != nil {
		return 0, 0, fmt.Errorf("generate key pair: %w", err)
	}
	return pub, priv, nil
}
