package main

import (
	"fmt"
	"strings"

	"github.com/miekg/pkcs11"
)

// findSlot returns the slot that has a token with the given label.
func findSlot(p *pkcs11.Ctx, label string) (uint, error) {
	slots, err := p.GetSlotList(true)
	if err != nil {
		return 0, fmt.Errorf("list slots: %w", err)
	}
	for _, sl := range slots {
		info, err := p.GetTokenInfo(sl)
		if err == nil && strings.TrimSpace(info.Label) == label {
			return sl, nil
		}
	}
	return 0, fmt.Errorf("token %q not found", label)
}
