package broadcast

import (
	"fmt"
	"log"
	"testing"
)

func TestBroadcast(t *testing.T) {
	n := 10
	cpk, secretKeys, err := Setup(n)
	if err != nil {
		log.Fatalln(err)
	}

	S := []int{0, 1, 2, 3, 4}
	bpk := cpk.broadcastPublicKey()
	hdr, K, err := bpk.Encrypt(S)
	if err != nil {
		log.Fatalln(err)
	}

	chkK := secretKeys[0].Decrypt(S, hdr, cpk.getPublicKey(0)).Marshal()
	if string(K.Marshal()) != string(chkK) {
		fmt.Printf("Equality check failed\nK: %v\nchkK: %v", K.Marshal(), chkK)
	}
	fmt.Printf("Equality check success\n")
}
