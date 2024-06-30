package broadcast

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fentec-project/bn256"
	"math/big"
)

type CompletePublicKey struct {
	PArr      []bn256.G1 `json:"g1"`
	QArr      []bn256.G2 `json:"g2"`
	V         bn256.G1   `json:"v"`
	GroupName string     `json:"name"`
}

func (c *CompletePublicKey) MarshalJSON() ([]byte, error) {
	pArr := make([]string, len(c.PArr))
	for i, p := range c.PArr {
		pArr[i] = fmt.Sprintf("%x", p.Marshal())
	}

	qArr := make([]string, len(c.QArr))
	for i, q := range c.QArr {
		qArr[i] = fmt.Sprintf("%x", q.Marshal())
	}

	v := fmt.Sprintf("%x", c.V.Marshal())

	return json.Marshal(&struct {
		G1   []string `json:"g1"`
		G2   []string `json:"g2"`
		V    string   `json:"v"`
		Name string   `json:"name"`
	}{
		G1:   pArr,
		G2:   qArr,
		V:    v,
		Name: c.GroupName,
	})
}

func (c *CompletePublicKey) UnmarshalJSON(data []byte) error {
	var aux struct {
		G1   []string `json:"g1"`
		G2   []string `json:"g2"`
		V    string   `json:"v"`
		Name string   `json:"name"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.PArr = make([]bn256.G1, len(aux.G1))
	for i, pStr := range aux.G1 {
		pBytes, err := hex.DecodeString(pStr)
		if err != nil {
			return err
		}
		_, err = c.PArr[i].Unmarshal(pBytes)
		if err != nil {
			return err
		}
	}

	c.QArr = make([]bn256.G2, len(aux.G2))
	for i, qStr := range aux.G2 {
		qBytes, err := hex.DecodeString(qStr)
		if err != nil {
			return err
		}
		_, err = c.QArr[i].Unmarshal(qBytes)
		if err != nil {
			return err
		}
	}

	vBytes, err := hex.DecodeString(aux.V)
	if err != nil {
		return err
	}
	_, err = c.V.Unmarshal(vBytes)
	if err != nil {
		return err
	}

	c.GroupName = aux.Name

	return nil
}
func (pk *CompletePublicKey) String() string {
	//fmt.Println(pk.PArr[0].String())
	//fmt.Println(pk.QArr[0].String())
	jsonData, _ := json.Marshal(pk)
	return string(jsonData)
}

func JSON2CompletePublicKey(cpkStr string) CompletePublicKey {
	var newPk CompletePublicKey
	_ = json.Unmarshal(([]byte)(cpkStr), &newPk)
	//fmt.Println("反序列化成功", newPk.QArr[1].String())
	return newPk
}

type Header struct {
	C0  *bn256.G1
	C0p *bn256.G2
	C1  *bn256.G1
}

type AdvertiserSecretKey struct {
	I  int
	Di bn256.G1
}

func Setup(n int, name string) (CompletePublicKey, []AdvertiserSecretKey) {
	r := rand.Reader
	//_, P, _ := bn256.RandomG1(r)
	//_, Q, _ := bn256.RandomG2(r)
	P := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	Q := new(bn256.G2).ScalarBaseMult(big.NewInt(1))
	alpha, _ := rand.Int(r, bn256.Order)
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
		//fmt.Println(I, PArr[I].String())
	}
	for i := 1; i < n+1; i++ {
		accumulatorQ = accumulatorQ.ScalarMult(accumulatorQ, alpha)
		QArr[i] = *new(bn256.G2).Set(accumulatorQ)
		//fmt.Println(I, QArr[I].String())
	}

	gamma, _ := rand.Int(r, bn256.Order)
	//gamma = big.NewInt(2) //
	V := new(bn256.G1).ScalarMult(P, gamma)
	//fmt.Println(V.String() == PArr[1].String())
	privateKeys := make([]AdvertiserSecretKey, n+1)
	for i := 0; i < n+1; i++ {
		privateKeys[i] = AdvertiserSecretKey{
			I:  i,
			Di: *new(bn256.G1).ScalarMult(&PArr[i], gamma),
		}
	}

	return CompletePublicKey{
		PArr:      PArr,
		QArr:      QArr,
		V:         *V,
		GroupName: name,
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
	i := adsk.I
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
