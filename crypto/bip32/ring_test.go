package bip32

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	ring "github.com/noot/ring-go"
	"golang.org/x/crypto/sha3"
)

func signAndVerify(curve ring.Curve) {
	privkey := curve.NewRandomScalar()
	data := make([]byte, 1024) // 1KB
	_, err := rand.Read(data)
	msgHash := sha3.Sum256(data)

	// size of the public key ring (anonymity set)
	const size = 15

	// our key's secret index within the set
	const idx = 2

	keyring, err := ring.NewKeyRing(curve, size, privkey, idx)
	if err != nil {
		panic(err)
	}
	starttime := time.Now().UnixMicro()
	sig, err := keyring.Sign(msgHash, privkey)
	endtime := time.Now().UnixMicro()
	fmt.Printf("sign time cost %d us\n", (endtime - starttime))
	if err != nil {
		panic(err)
	}

	ok := sig.Verify(msgHash)
	if !ok {
		fmt.Println("failed to verify :(")
		return
	}

	fmt.Println("verified signature!")
}

func TestRingSig(t *testing.T) {
	fmt.Println("using secp256k1...")
	signAndVerify(ring.Secp256k1())
	//fmt.Println("using ed25519...")
	//signAndVerify(ring.Ed25519())
}
