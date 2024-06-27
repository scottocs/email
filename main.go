package main

import (
	"crypto/rand"
	"email/compile/contract"
	"email/crypto/aes"
	"email/crypto/stealth"
	"email/utils"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fentec-project/bn256"
	shell "github.com/ipfs/go-ipfs-api"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
)

func main() {

	contract_name := "Email"
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	//Bob's
	privatekey1 := utils.GetENV("PRIVATE_KEY_1")
	//Alice's
	privatekey2 := utils.GetENV("PRIVATE_KEY_2")
	deployTX := utils.Transact(client, privatekey1, big.NewInt(0), nil, nil)
	address, _ := utils.Deploy(client, contract_name, deployTX.(*bind.TransactOpts))
	ctc, err := contract.NewContract(common.HexToAddress(address.Hex()), client)

	//Alice uploads her public keys (A B)
	ska, _ := rand.Int(rand.Reader, bn256.Order)
	skb, _ := rand.Int(rand.Reader, bn256.Order)
	A := new(bn256.G1).ScalarBaseMult(ska)
	B := new(bn256.G1).ScalarBaseMult(skb)
	//fmt.Println("Alice", ska)
	para := []interface{}{"UploadPK", "Alice", contract.EmailPK{utils.G1ToPoint(A), utils.G1ToPoint(B)}}
	receipt := utils.Transact(client, privatekey2, big.NewInt(0), ctc, para).(*types.Receipt)
	fmt.Printf("%v Gas used: %d\n", para[0], receipt.GasUsed)

	//Bob generates Alice's Stealth address after downloading Alice's public keys
	pkRes, _ := ctc.DownloadPK(&bind.CallOpts{}, "Alice")
	Ap := utils.PointToG1(pkRes.A)
	Bp := utils.PointToG1(pkRes.B)
	if A.String() != Ap.String() || B.String() != Bp.String() {
		return
	}
	sa := stealth.CalculatePub(stealth.PublicKey{Ap, Bp})

	r, _ := rand.Int(rand.Reader, bn256.Order)
	m, _ := rand.Int(rand.Reader, bn256.Order)
	key := new(bn256.G1).ScalarBaseMult(m)
	c1 := new(bn256.G1).ScalarBaseMult(r) // c1 = r * G
	c2 := new(bn256.G1).Add(new(bn256.G1).ScalarMult(sa.S, r), key)

	msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00")
	ct, _ := aes.Encrypt(msg, key.Marshal()[:32])
	fmt.Println("encrypted email:", ct)
	// Bob uploads encrypted email content to IPFS
	sh := shell.NewShell("localhost:5001")
	cid, _ := sh.Add(strings.NewReader(ct))
	fmt.Println("send mail IPFS link:", cid)

	//Bob initiates BIP32 wallet address, Master secret key PRIVATE_KEY_100, 10 child keys PRIVATE_KEY_[11-20]
	utils.InitBIP32Wallet(client, privatekey1)
	privatekey11 := utils.GetENV("PRIVATE_KEY_11")
	para = []interface{}{"MailTo",
		contract.EmailStealthAddrPub{utils.G1ToPoint(sa.R), utils.G1ToPoint(sa.S)},
		contract.EmailElGamalCT{utils.G1ToPoint(c1), utils.G1ToPoint(c2)},
		cid, []string{"Alice", "Bob"}}
	receipt = utils.Transact(client, privatekey11, big.NewInt(0), ctc, para).(*types.Receipt)
	fmt.Printf("%v Gas used: %d\n", para[0], receipt.GasUsed)

	//Alice downloads the encrypted email
	mailRes, _ := ctc.DownloadMail(&bind.CallOpts{}, cid)
	// Alice obtains sp i.e., logS
	sp := stealth.ResolvePriv(stealth.SecretKey{ska, skb},
		stealth.StealthAddrPub{utils.PointToG1(mailRes.Pub.R), utils.PointToG1(mailRes.Pub.S)})
	c1pNeg := new(bn256.G1).Neg(utils.PointToG1(mailRes.Ct.C1))
	c2p := utils.PointToG1(mailRes.Ct.C2)
	keyp := new(bn256.G1).Add(c2p, new(bn256.G1).ScalarMult(c1pNeg, sp))
	//fmt.Println(key, keyp)
	if key.String() != keyp.String() {
		fmt.Println("Alice is unable to get the aes decryption key")
		return
	}
	sh.Get(cid, "./alice/")
	file, err := os.Open("./alice/" + cid)
	content, _ := io.ReadAll(file)
	decRes, _ := aes.Decrypt(string(content), keyp.Marshal()[:32])
	fmt.Println("Email content:", decRes)
}
