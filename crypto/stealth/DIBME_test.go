package stealth

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/fentec-project/bn256"
)

// -------------------------------
// 工具函数
// -------------------------------
func hashToG1(id string) *bn256.G1 {
	h := sha256.Sum256([]byte(id))
	scalar := new(big.Int).SetBytes(h[:])
	return new(bn256.G1).ScalarBaseMult(scalar)
}

func hashToG2(id string) *bn256.G2 {
	h := sha256.Sum256([]byte(id))
	scalar := new(big.Int).SetBytes(h[:])
	return new(bn256.G2).ScalarBaseMult(scalar)
}

func randScalar() *big.Int {
	r, _ := rand.Int(rand.Reader, bn256.Order)
	return r
}

func xorBytes(data, mask []byte) []byte {
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ mask[i%len(mask)]
	}
	return out
}

// -------------------------------
// DIB-ME 普通模式
// -------------------------------
func Setup() (*bn256.G1, *bn256.G1, *bn256.G1, [3]*big.Int) {
	g := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	a := randScalar()
	b := randScalar()
	c := randScalar()
	G := new(bn256.G1).ScalarMult(g, a)
	G_ := new(bn256.G1).ScalarMult(g, c)
	return g, G, G_, [3]*big.Int{a, b, c}
}

func SGen(sk [3]*big.Int, ids string) *bn256.G1 {
	b := sk[1]
	hs := hashToG1(ids)
	return new(bn256.G1).ScalarMult(hs, b)
}

func RGen(sk [3]*big.Int, idr string) (*bn256.G2, *bn256.G2) {
	a, b := sk[0], sk[1]
	hr := hashToG2(idr)
	dkr1 := new(bn256.G2).ScalarMult(hr, a)
	dkr2 := new(bn256.G2).ScalarMult(hr, b)
	return dkr1, dkr2
}

func Enc(g, G *bn256.G1, eks *bn256.G1, idr string, message []byte) (*bn256.G1, *bn256.G1, []byte) {
	v := randScalar()
	U := new(bn256.G1).ScalarBaseMult(randScalar())
	V := new(bn256.G1).ScalarMult(g, v)

	hr := hashToG2(idr)
	k11 := bn256.Pair(new(bn256.G1).ScalarMult(G, v), hr)
	k12 := bn256.Pair(new(bn256.G1).Add(U, eks), hr)

	temp := xorBytes(k11.Marshal(), k12.Marshal())
	k1 := sha256.Sum256(temp)

	mask := k1[:]
	C3 := xorBytes(message, mask)

	return U, V, C3
}

func Dec(dkr1, dkr2 *bn256.G2, R, idS string, U, V *bn256.G1, C3 []byte) []byte {
	k11 := bn256.Pair(V, dkr1)
	k12 := new(bn256.GT).Add(
		bn256.Pair(U, hashToG2(R)),
		bn256.Pair(hashToG1(idS), dkr2),
	)

	temp := xorBytes(k11.Marshal(), k12.Marshal())
	k1 := sha256.Sum256(temp)

	mask := k1[:]
	message := xorBytes(C3, mask)
	return message
}

// -------------------------------
// 测试函数：含时间开销
// -------------------------------
func TestDIBME(t *testing.T) {
	g, G, _, sk := Setup()
	eks := SGen(sk, "alice")
	dkr1, dkr2 := RGen(sk, "bob")

	msg := []byte("Secret Message in Go")

	// 测试加密时间
	startEnc := time.Now()
	U, V, C3 := Enc(g, G, eks, "bob", msg)
	encDuration := time.Since(startEnc)

	// 测试解密时间
	startDec := time.Now()
	recovered := Dec(dkr1, dkr2, "bob", "alice", U, V, C3)
	decDuration := time.Since(startDec)

	fmt.Println("Original:", string(msg))
	fmt.Println("Recovered:", string(recovered))
	fmt.Printf("Enc time: %v\n", encDuration)
	fmt.Printf("Dec time: %v\n", decDuration)

	if string(msg) != string(recovered) {
		t.Errorf("decryption failed: got %s, want %s", string(recovered), string(msg))
	}
}
