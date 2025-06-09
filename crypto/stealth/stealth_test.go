package stealth

import (
	"crypto/rand"
	"fmt"
	"github.com/fentec-project/bn256"
	"testing"
	"time"
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
	stealthPriv := ResolvePriv(priv, *stealthPub)
	fmt.Println(stealthPub.S)
	fmt.Println(new(bn256.G1).ScalarBaseMult(stealthPriv))
	var n int64 = 1000
	starttime := time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		stealthPub = CalculatePub(pub)
	}
	endtime := time.Now().UnixMicro()
	fmt.Printf("CalculatePub time cost %d us\n", (endtime-starttime)/n)
	starttime = time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		stealthPriv = ResolvePriv(priv, *stealthPub)
	}
	endtime = time.Now().UnixMicro()
	fmt.Printf("ResolvePriv time cost %d us\n", (endtime-starttime)/n)
}
