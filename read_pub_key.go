package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/miekg/pkcs11"
)

// readPublicKey reads the public EC point from the HSM and converts it
// to a Go *ecdsa.PublicKey.
func readPublicKey(p *pkcs11.Ctx, s pkcs11.SessionHandle, obj pkcs11.ObjectHandle) (*ecdsa.PublicKey, error) {
	attrs, err := p.GetAttributeValue(s, obj,
		[]*pkcs11.Attribute{pkcs11.NewAttribute(pkcs11.CKA_EC_POINT, nil)})
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	// The HSM returns the point wrapped in DER (an OCTET STRING).
	var point []byte
	if _, err := asn1.Unmarshal(attrs[0].Value, &point); err != nil {
		return nil, fmt.Errorf("unwrap EC point: %w", err)
	}

	// point = 0x04 || X (48 bytes) || Y (48 bytes) = 97 bytes
	if len(point) != 1+2*coordSize || point[0] != 0x04 {
		return nil, fmt.Errorf("unexpected EC point format (len=%d)", len(point))
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P384(),
		X:     new(big.Int).SetBytes(point[1 : 1+coordSize]),
		Y:     new(big.Int).SetBytes(point[1+coordSize:]),
	}, nil
}
