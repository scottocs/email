package stealth

import (
	"crypto/rand"
	"fmt"
	"github.com/fentec-project/bn256"
	"testing"
)

func TestStealth(t *testing.T) {

	a, _ := rand.Int(rand.Reader, bn256.Order)
	b, _ := rand.Int(rand.Reader, bn256.Order)

	priv := SecretKey{
		a,
		b,
	}
	pub := PublicKey{
		new(bn256.G1).ScalarBaseMult(a),
		new(bn256.G1).ScalarBaseMult(b),
	}
	stealthPub := CalculatePub(pub)
	stealthPriv := ResolveSec(priv, stealthPub)
	fmt.Println(stealthPub.S)
	fmt.Println(new(bn256.G1).ScalarBaseMult(stealthPriv))
}
