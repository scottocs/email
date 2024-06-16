package main

import (
	"context"
	"crypto/rand"
	"email/compile/contract"
	"email/crypto/rwdabe"
	"email/utils"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	bn128 "github.com/fentec-project/bn256"
	lib "github.com/fentec-project/gofe/abe"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"strconv"
)

func randomInt(curveOrder *big.Int) *big.Int {
	// Generate a random number in [0, curve_order-1]
	n, err := rand.Int(rand.Reader, curveOrder)
	if err != nil {
		panic(err)
	}
	// Add 1 to shift to [1, curve_order]
	n.Add(n, big.NewInt(1))
	return n
}

// keccak256 computes the Keccak256 hash of the input data.
func keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// hash2G1 hashes the input data to a point on the bn256 G1 curve.
func hash2G1(data []byte) *bn128.G1 {
	hash := keccak256(data)
	intHash := new(big.Int).SetBytes(hash)
	g1 := new(bn128.G1)
	g1.ScalarBaseMult(intHash)
	return g1
}

// hash2int converts the Keccak256 hash of the input data to an integer.
func hash2int(data []byte) *big.Int {
	hash := keccak256(data)
	intHash := new(big.Int).SetBytes(hash)
	return intHash
}

func testHash2G1() {
	data := []byte("data")

	// Compute Keccak256 hash
	hash := keccak256(data)
	fmt.Println("Keccak256 hash:", hex.EncodeToString(hash))

	// Convert hash to an integer
	intHash := new(big.Int).SetBytes(hash)
	fmt.Println("Keccak256 hash as int:", intHash)

	// Hash to G1
	g1Point := hash2G1(data)
	fmt.Println("G1 Point:", g1Point)
}
func generateACPStr(n int) string {

	attrs := make([]string, n)
	for i := 1; i < n+1; i++ {
		attrs[i-1] = "auth" + strconv.Itoa(i) + ":at1"
		//fmt.Println(strconv.Itoa(i))
	}

	acp := utils.RandomACP(attrs)
	//fmt.Println(acp, attrs)
	return acp
}

func main() {

	contract_name := "Email"

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	privatekey1 := utils.GetENV("PRIVATE_KEY_1")

	deployTX := utils.Transact(client, privatekey1, big.NewInt(0))

	address, _ := utils.Deploy(client, contract_name, deployTX)

	ctc, err := contract.NewContract(common.HexToAddress(address.Hex()), client)

	auth0 := utils.Transact(client, privatekey1, big.NewInt(0))
	tx3, err := ctc.HashToG1(auth0, "11223")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Onchain HashToG1() result:", tx3)
	receipt3, _ := bind.WaitMined(context.Background(), client, tx3)
	fmt.Printf("HashToG1() Gas used: %d\n", receipt3.GasUsed)

	fmt.Println("...........................................................Setup............................................................")

	maabe := rwdabe.NewMAABE()

	// create  authorities, each with two attributes
	attribs1 := []string{"auth1:at1", "auth1:at2"}
	attribs2 := []string{"auth2:at1", "auth2:at2"}
	attribs3 := []string{"auth3:at1", "auth3:at2"}
	attribs4 := []string{"auth4:at1", "auth4:at2"}
	auth1, _ := maabe.NewMAABEAuth("auth1")
	auth2, _ := maabe.NewMAABEAuth("auth2")
	auth3, _ := maabe.NewMAABEAuth("auth3")
	auth4, _ := maabe.NewMAABEAuth("auth4")

	fmt.Println("..........................................................Encryption...........................................................")

	// acjudges := make([]*ACJudge, 1)
	acp := generateACPStr(4)
	fmt.Println("Access Control Policy:", acp)
	//msp, _ := lib.BooleanToMSP("auth1:at1 AND auth2:at1 AND auth3:at1 AND auth3:at2 AND auth4:at1", false)
	msp, _ := lib.BooleanToMSP(acp, false)
	// define the set of all public keys we use
	pks := []*rwdabe.MAABEPubKey{auth1.Pk, auth2.Pk, auth3.Pk, auth4.Pk}

	// choose a message
	msg := "Attack at dawn!"

	// encrypt the message with the decryption policy in msp
	ct, _ := maabe.ABEEncrypt(msg, msp, pks)

	fmt.Println("..........................................................Key Generation...........................................................")

	// choose a single user's Global ID
	gid := "gid1"
	// authority 1 issues keys to user
	key11, _ := auth1.ABEKeyGen(gid, attribs1[0])
	key12, _ := auth1.ABEKeyGen(gid, attribs1[1])
	// authority 2 issues keys to user
	key21, _ := auth2.ABEKeyGen(gid, attribs2[0])
	key22, _ := auth2.ABEKeyGen(gid, attribs2[1])
	// authority 4 issues keys to user
	key31, _ := auth3.ABEKeyGen(gid, attribs3[0])
	key32, _ := auth3.ABEKeyGen(gid, attribs3[1])

	// authority 4 issues keys to user
	key41, _ := auth4.ABEKeyGen(gid, attribs4[0])
	key42, _ := auth4.ABEKeyGen(gid, attribs4[1])

	// user tries to decrypt with different key combos
	ks1 := []*rwdabe.MAABEKey{key11, key21, key31, key41} // ok
	ks2 := []*rwdabe.MAABEKey{key12, key22, key32}        // ok
	ks5 := []*rwdabe.MAABEKey{key31, key32, key42}        // ok

	fmt.Println("..........................................................Decryption...........................................................")
	// try to decrypt all messages
	msg1, _ := maabe.ABEDecrypt(ct, ks1)

	msg2, _ := maabe.ABEDecrypt(ct, ks2)

	msg5, _ := maabe.ABEDecrypt(ct, ks5)

	fmt.Println("msg1", msg1)
	fmt.Println("msg2", msg2)
	fmt.Println("msg5", msg5)
	fmt.Println("Decrypt Result is:", msg1 == msg)

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
