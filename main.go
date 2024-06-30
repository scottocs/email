package main

import (
	"crypto/rand"
	"email/compile/contract"
	"email/crypto/broadcast"
	"email/utils"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fentec-project/bn256"
	"log"
	"math/big"
	"strconv"
)

func main() {

	contract_name := "Email"
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	//Deploy
	deployTX := utils.Transact(client, utils.GetENV("PRIVATE_KEY_1"), big.NewInt(0), nil, nil)
	address, _ := utils.Deploy(client, contract_name, deployTX.(*bind.TransactOpts))
	ctc, err := contract.NewContract(common.HexToAddress(address.Hex()), client)

	fmt.Println("=============================upload personal public keys=====================")
	//Users register their public keys (A B)
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	users := make([]utils.User, len(names))
	for i := 0; i < len(names); i++ {
		a, _ := rand.Int(rand.Reader, bn256.Order)
		b, _ := rand.Int(rand.Reader, bn256.Order)
		A := new(bn256.G1).ScalarBaseMult(a)
		B := new(bn256.G1).ScalarBaseMult(b)
		para := []interface{}{"Register", names[i], contract.EmailPK{utils.G1ToPoint(A), utils.G1ToPoint(B)}}
		privatekey := utils.GetENV("PRIVATE_KEY_" + strconv.Itoa(i+1))
		_ = utils.Transact(client, privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
		users[i] = utils.User{names[i], a, b, A, B, privatekey, nil}
	}
	//users generate their BIP32 child keys
	utils.InitBIP32Wallet(client, users)

	Bob := users[0]
	Alice := users[1]
	//Charlie := users[2]
	//Bob generates Alice's Stealth address after downloading Alice's public keys
	fmt.Println("=============================test one-to-one mailing=====================")
	m, _ := rand.Int(rand.Reader, bn256.Order)
	key := new(bn256.G1).ScalarBaseMult(m)
	msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00. \nBest,\nBob")
	recs := []string{names[8], names[9]} //user names to confuse others
	utils.MailTo(client, ctc, Bob, key, msg, Alice, recs)
	//Alice downloads the encrypted email
	utils.DownloadMail(ctc, Alice)

	//	broadcast encryption
	//Bob is a group administrator
	n := len(users)
	cpk, secretKeys := broadcast.Setup(n, "@group1")
	Bobindex := 0
	Bob = users[Bobindex]
	Bob.BrdUser = &utils.DomainUser{utils.Group{cpk, secretKeys[Bobindex]}, nil, nil}
	fmt.Println("=============================upload broadcast public keys=====================")
	//Bob sends secretKeys to each group member via one-to-one mailing (Here, sends secretkeys to the domain)
	utils.RegisterGroup(client, ctc, Bob, cpk, secretKeys, users)

	fmt.Println("=============================Alice Download broadcast public keys=====================")
	//brdPriv := utils.DownloadAndResolvePriv(ctc, Alice, cpk.GroupName)

	//
	////Charlie is a domain manager, Charlie generates \prod_j∈S g_{n+1-j} for the domain
	fmt.Println("=============================build domain public keys=====================")
	size := n / 2
	S := make([]int, size)
	for i := 0; i < size; i++ {
		S[i] = i + 1
	}
	domainPK := cpk.BuildDomainPK(S)
	Charlieindex := 2
	Charlie := users[Charlieindex]
	////TODO obtain from blockchain
	Charlie.BrdUser = &utils.DomainUser{utils.Group{cpk, secretKeys[Bobindex]}, S, domainPK}
	//senderSK := secretKeys[Charlieindex].Di
	//// todo Charlie sends domainPK to Bob via one-to-one mailing
	//Bob encrypts a mail content and broadcast to the domain
	msg = []byte("Dear Staff, we are going to have a meeting at Jun 30, 2024 09:00 at the gym. \nBest,\nDomain manager")
	//cid := utils.MailTo(client, ctc, Bob, key, msg, Alice, recs)
	cid2 := utils.BroadcastTo(client, ctc, Charlie, cpk, msg)
	//pairingRes, _ := ctc.GetPairingRes(&bind.CallOpts{})
	//fmt.Printf("GetPairingRes: %v\n", pairingRes)
	//
	////todo Alice downloads mails and decrypt cid2,GroupName
	//
	brdPriv := utils.DownloadAndResolvePriv(ctc, Alice, cpk.GroupName)
	Alice.BrdUser = &utils.DomainUser{utils.Group{cpk, brdPriv}, S, domainPK}
	utils.DownloadBrdMail(ctc, Alice, cpk, brdPriv, cid2)

}
