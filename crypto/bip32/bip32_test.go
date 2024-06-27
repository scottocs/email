package bip32

import (
	"fmt"
	"github.com/tyler-smith/go-bip32"
	"log"
	"testing"
)

func TestBIP32(t *testing.T) {
	// test BIP32
	seed, err := bip32.NewSeed()
	if err != nil {
		log.Fatalln("Error generating seed:", err)
	}

	// Create master private key from seed
	computerVoiceMasterKey, _ := bip32.NewMasterKey(seed)

	// Map departments to keys
	// There is a very small chance a given child index is invalid
	// If so your real program should handle this by skipping the index
	departmentKeys := map[string]*bip32.Key{}
	departmentKeys["Sales"], _ = computerVoiceMasterKey.NewChildKey(0)
	departmentKeys["Marketing"], _ = computerVoiceMasterKey.NewChildKey(1)
	departmentKeys["Engineering"], _ = computerVoiceMasterKey.NewChildKey(2)
	departmentKeys["Customer Support"], _ = computerVoiceMasterKey.NewChildKey(3)

	// Create public keys for record keeping, auditors, payroll, etc
	departmentAuditKeys := map[string]*bip32.Key{}
	departmentAuditKeys["Sales"] = departmentKeys["Sales"].PublicKey()
	departmentAuditKeys["Marketing"] = departmentKeys["Marketing"].PublicKey()
	departmentAuditKeys["Engineering"] = departmentKeys["Engineering"].PublicKey()
	departmentAuditKeys["Customer Support"] = departmentKeys["Customer Support"].PublicKey()

	// Print public keys
	for department, pubKey := range departmentAuditKeys {
		fmt.Println(department, pubKey)
	}

}
