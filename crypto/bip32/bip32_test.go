package bip32

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/tyler-smith/go-bip32"
)

// 将 BIP32 子密钥转换为 ECDSA 私钥
func bip32ToECDSA(key *bip32.Key) *ecdsa.PrivateKey {
	curve := elliptic.P256()
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = new(big.Int).SetBytes(key.Key)
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(priv.D.Bytes())
	return priv
}

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
	var n int64 = 1000
	starttime := time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		departmentKeys["Sales"], _ = computerVoiceMasterKey.NewChildKey(11111111)
	}

	ecdsaKey := bip32ToECDSA(departmentKeys["Sales"])
	data := make([]byte, 1024) // 1KB
	rand.Read(data)
	hash := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	valid := ecdsa.Verify(&ecdsaKey.PublicKey, hash[:], r, s)

	//signDuration := time.Since(startSign)
	endtime := time.Now().UnixMicro()
	fmt.Printf("NewChildKey time cost %d us\n valid: %v\n", (endtime-starttime)/n, valid)
	departmentKeys["Marketing"], _ = computerVoiceMasterKey.NewChildKey(1)
	departmentKeys["Engineering"], _ = computerVoiceMasterKey.NewChildKey(2)
	departmentKeys["Customer Support"], _ = computerVoiceMasterKey.NewChildKey(3)

	// Create public keys for record keeping, auditors, payroll, etc
	departmentAuditKeys := map[string]*bip32.Key{}
	starttime = time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		departmentAuditKeys["Sales"] = departmentKeys["Sales"].PublicKey()
	}
	endtime = time.Now().UnixMicro()
	fmt.Printf("PublicKey() time cost %d us\n", (endtime-starttime)/n)
	departmentAuditKeys["Marketing"] = departmentKeys["Marketing"].PublicKey()
	departmentAuditKeys["Engineering"] = departmentKeys["Engineering"].PublicKey()
	departmentAuditKeys["Customer Support"] = departmentKeys["Customer Support"].PublicKey()

	// Print public keys
	for department, pubKey := range departmentAuditKeys {
		fmt.Println(department, pubKey)
	}

}
