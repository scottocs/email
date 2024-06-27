package stealth

//TODO refer to https://github.com/liyue201/stealth-address-demo/blob/main/main.go i.e., teststealth.go
// or refer to https://github.com/hacheigriega/go-stealth/blob/main/stealth.go

import (
	"crypto/rand"
	"fmt"
	"github.com/fentec-project/bn256"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type PublicKey struct {
	A *bn256.G1
	B *bn256.G1
}
type SecretKey struct {
	a *big.Int
	b *big.Int
}
type StealthAddrPub struct {
	R *bn256.G1
	S *bn256.G1
}

func Hash2Int(msg string) *big.Int {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(msg))
	v := hash.Sum(nil)
	return new(big.Int).SetBytes(v)

}

func CalculatePub(pub PublicKey) *StealthAddrPub {
	r, _ := rand.Int(rand.Reader, bn256.Order)
	g := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	R := new(bn256.G1).ScalarMult(g, r)
	//fmt.Println(g)
	hash := Hash2Int(new(bn256.G1).ScalarMult(pub.B, r).String())
	//fmt.Println(h)
	return &StealthAddrPub{R,
		new(bn256.G1).Add(pub.A, new(bn256.G1).ScalarMult(g, hash))}
}

func ResolveSec(priv SecretKey, stealth *StealthAddrPub) *big.Int {

	h := Hash2Int(new(bn256.G1).ScalarMult(stealth.R, priv.b).String())
	s := new(big.Int).Add(priv.a, h)
	Sp := new(bn256.G1).ScalarBaseMult(s)
	//fmt.Println(stealth.S)
	//fmt.Println(Sp)

	if stealth.S.String() != Sp.String() {
		fmt.Println("S is wrong, no secret is geneated")
		return nil
	}
	return s
}
