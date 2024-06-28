package main

import (
	"crypto/rand"
	"email/compile/contract"
	"email/crypto/aes"
	"email/crypto/broadcast"
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
	_ = utils.Transact(client, privatekey2, big.NewInt(0), ctc, para).(*types.Receipt)
	//fmt.Printf("%v Gas used: %d\n", para[0], receipt.GasUsed)

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

	msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00. \nBest,\nBob")
	ct, _ := aes.Encrypt(msg, key.Marshal()[:32])
	fmt.Println("encrypted email:", ct)
	// Bob uploads encrypted email content to IPFS
	sh := shell.NewShell("localhost:5001")
	cid, _ := sh.Add(strings.NewReader(ct))
	fmt.Println("One-to-one mail IPFS link:", cid)

	//Bob initiates BIP32 wallet address, Master secret key PRIVATE_KEY_100, 10 child keys PRIVATE_KEY_[11-20]
	utils.InitBIP32Wallet(client, privatekey1)
	privatekey11 := utils.GetENV("PRIVATE_KEY_11")
	para = []interface{}{"MailTo",
		contract.EmailStealthAddrPub{utils.G1ToPoint(sa.R), utils.G1ToPoint(sa.S)},
		contract.EmailElGamalCT{utils.G1ToPoint(c1), utils.G1ToPoint(c2)},
		cid, []string{"Alice", "Bob"}}
	_ = utils.Transact(client, privatekey11, big.NewInt(0), ctc, para).(*types.Receipt)

	//Alice downloads the encrypted email
	mailRes, _ := ctc.DownloadMail(&bind.CallOpts{}, cid)
	// Alice obtains sp i.e., logS
	sp := stealth.ResolvePriv(stealth.SecretKey{ska, skb},
		stealth.StealthAddrPub{utils.PointToG1(mailRes.Pub.R), utils.PointToG1(mailRes.Pub.S)})
	c1pNeg := new(bn256.G1).Neg(utils.PointToG1(mailRes.Ct.C1))
	c2p := utils.PointToG1(mailRes.Ct.C2)
	keyp := new(bn256.G1).Add(c2p, new(bn256.G1).ScalarMult(c1pNeg, sp))
	sh.Get(cid, "./alice/")
	file, err := os.Open("./alice/" + cid)
	content, _ := io.ReadAll(file)
	decRes, _ := aes.Decrypt(string(content), keyp.Marshal()[:32])
	fmt.Println("One-to-one Email content:", decRes)

	//	broadcast encryption
	//Charlie is a group administrator
	n := 10
	cpk, secretKeys := broadcast.Setup(n)

	// todo Charlie sends secretKeys to each group member via one-to-one mailing

	//Bob is a domain manager, Charlie generates \prod_j∈S g_{n+1-j} for the domain
	S := []int{1, 3, 8, 6}
	senderIndex := S[0]
	domainPK := cpk.BuildDomainPK(S)
	hdr, beK := cpk.Encrypt(domainPK)
	senderSK := secretKeys[senderIndex].Di
	senderPK := cpk.PArr[senderIndex]
	x, _ := rand.Int(rand.Reader, bn256.Order)
	domainRecivers := contract.EmailBrdcastCT{utils.G1ToPoint(hdr.C0), utils.G1ToPoint(hdr.C1)}
	proof := contract.EmailDomainProof{utils.G1ToPoint(senderSK.ScalarBaseMult(x)),
		utils.G1ToPoint(&senderPK), utils.G1ToPoint(cpk.V.ScalarBaseMult(x))}
	// todo Charlie sends domainPK to Bob via one-to-one mailing
	//
	//Bob encrypts a mail content and broadcast to the domain
	msg = []byte("Dear Staff, we are going to have a meeting at Jun 30, 2024 09:00 at the gym. \nBest,\nDomain manager")
	ct, _ = aes.Encrypt(msg, beK.Marshal()[:32])
	fmt.Println("encrypted email to broadcast:", ct)
	// Bob uploads encrypted email content to IPFS
	cid2, _ := sh.Add(strings.NewReader(ct))
	fmt.Println("broadcast mail IPFS link:", cid2)
	privatekey12 := utils.GetENV("PRIVATE_KEY_12")
	para = []interface{}{"BrdcastTo", domainRecivers, proof, cid2, []string{"Alice", "Bob"}}
	_ = utils.Transact(client, privatekey12, big.NewInt(0), ctc, para).(*types.Receipt)
	decIndex := S[1]
	beKp := secretKeys[decIndex].Decrypt(S, hdr, cpk)
	sh.Get(cid2, "./alice/")
	file, _ = os.Open("./alice/" + cid2)
	content, _ = io.ReadAll(file)
	decRes, _ = aes.Decrypt(string(content), beKp.Marshal()[:32])
	fmt.Println("Broadcast Email content:", decRes)

}
