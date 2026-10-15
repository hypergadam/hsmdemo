package main

import (
	"fmt"

	"github.com/miekg/pkcs11"
)

// findObject finds one object by class and label. It returns 0 if not found.
func findObject(p *pkcs11.Ctx, s pkcs11.SessionHandle, class uint, label string) (pkcs11.ObjectHandle, error) {
	tmpl := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, class),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
	}
	if err := p.FindObjectsInit(s, tmpl); err != nil {
		return 0, fmt.Errorf("find init: %w", err)
	}
	objs, _, err := p.FindObjects(s, 1)
	if err != nil {
		return 0, fmt.Errorf("find objects: %w", err)
	}
	if err := p.FindObjectsFinal(s); err != nil {
		return 0, fmt.Errorf("find final: %w", err)
	}
	if len(objs) == 0 {
		return 0, nil
	}
	return objs[0], nil
}
