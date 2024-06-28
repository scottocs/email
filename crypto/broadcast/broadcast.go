package broadcast

import (
	"crypto/rand"
	"fmt"
	"github.com/fentec-project/bn256"
	"math/big"
)

type CompletePublicKey struct {
	//n int
	//P    bn256.G1
	PArr []bn256.G1
	//Q    bn256.G2
	QArr []bn256.G2
	V    bn256.G1
}

//type BroadcastPublicKey struct {
//	//n    int
//	//P    bn256.G1
//	PArr []bn256.G1
//	//Q    bn256.G2
//	Q1 bn256.G2
//	V  bn256.G1
//}

type Header struct {
	C0  *bn256.G1
	C0p *bn256.G2
	C1  *bn256.G1
}

//type AdvertiserPublicKey struct {
//	N  int
//	Qi bn256.G2
//	//DomainPK1 bn256.G1
//	PArr []bn256.G1
//}

type AdvertiserSecretKey struct {
	i  int
	Di bn256.G1
}

func Setup(n int) (CompletePublicKey, []AdvertiserSecretKey) {
	r := rand.Reader
	//_, P, _ := bn256.RandomG1(r)
	//_, Q, _ := bn256.RandomG2(r)
	P := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	Q := new(bn256.G2).ScalarBaseMult(big.NewInt(1))
	//fmt.Println(P.String())
	//fmt.Println(Q.String())
	alpha, _ := rand.Int(r, bn256.Order)
	alpha = big.NewInt(2) //
	// build 2n-1 P_i values
	accumulatorP := new(bn256.G1).Set(P)
	accumulatorQ := new(bn256.G2).Set(Q)
	PArr := make([]bn256.G1, 2*n+1)
	QArr := make([]bn256.G2, n+1)
	PArr[0] = *new(bn256.G1).Set(P)
	QArr[0] = *new(bn256.G2).Set(Q)
	for i := 1; i < 2*n+1; i++ {
		accumulatorP = accumulatorP.ScalarMult(accumulatorP, alpha)
		PArr[i] = *new(bn256.G1).Set(accumulatorP)
		if i == n+1 {
			PArr[i] = *new(bn256.G1).ScalarBaseMult(big.NewInt(0))
		}
		//fmt.Println(i, PArr[i].String())
	}
	for i := 1; i < n+1; i++ {
		accumulatorQ = accumulatorQ.ScalarMult(accumulatorQ, alpha)
		QArr[i] = *new(bn256.G2).Set(accumulatorQ)
		//fmt.Println(i, QArr[i].String())
	}

	gamma, _ := rand.Int(r, bn256.Order)
	gamma = big.NewInt(2) //
	V := new(bn256.G1).ScalarMult(P, gamma)
	//fmt.Println(V.String() == PArr[1].String())
	privateKeys := make([]AdvertiserSecretKey, n+1)
	for i := 0; i < n+1; i++ {
		privateKeys[i] = AdvertiserSecretKey{
			i:  i,
			Di: *new(bn256.G1).ScalarMult(&PArr[i], gamma),
		}
		//fmt.Println(i, privateKeys[i].Di.String())
	}

	return CompletePublicKey{
		PArr: PArr,
		QArr: QArr,
		V:    *V,
	}, privateKeys
}

func (cpk *CompletePublicKey) BuildDomainPK(S []int) *bn256.G1 {
	n := len(cpk.QArr) - 1
	pk := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	for _, j := range S {
		pk = pk.Add(pk, &cpk.PArr[n+1-j])
	}
	//fmt.Println(pk)
	return pk
}
func (cpk *CompletePublicKey) Encrypt(domainPK *bn256.G1) (Header, *bn256.GT) {
	t, _ := rand.Int(rand.Reader, bn256.Order)
	//t = big.NewInt(2) //
	n := len(cpk.QArr) - 1
	g := &cpk.PArr[0]
	q := &cpk.QArr[0]

	hdr := Header{
		C0:  new(bn256.G1).ScalarMult(g, t),
		C0p: new(bn256.G2).ScalarMult(q, t),
		C1:  new(bn256.G1).ScalarMult(new(bn256.G1).Add(&cpk.V, domainPK), t),
	}
	ele := bn256.Pair(&cpk.PArr[n], &cpk.QArr[1])
	K := ele.ScalarMult(ele, t)
	return hdr, K
}

func (adsk *AdvertiserSecretKey) Decrypt(S []int, hdr Header, cpk CompletePublicKey) *bn256.GT {
	i := adsk.i
	if i == 0 {
		fmt.Println("Error: index 0 cannot be used")
	}
	n := len(cpk.QArr) - 1
	numerator := bn256.Pair(hdr.C1, &cpk.QArr[i])
	val := &adsk.Di
	for _, j := range S {
		if j != i {
			val = val.Add(val, &cpk.PArr[n+1-j+i])
		}
	}
	denominator := new(bn256.GT).Neg(bn256.Pair(val, hdr.C0p))
	out := new(bn256.GT).Add(numerator, denominator)

	return out
}
