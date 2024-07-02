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

	fmt.Println("=============================register personal public keys=====================")
	//Users register their public keys (A B)
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	//names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy", "Isabella", "Jacob", "Ava", "Matthew", "Mia", "Daniel", "Abigail", "Ethan", "Harper", "Alexander", "Amelia", "Ryan", "Evelyn", "Nathan", "Elizabeth", "Samuel", "Charlotte", "Christopher", "Grace", "Jonathan", "Lily", "Gabriel", "Ella", "Andrew", "Avery", "Joshua", "Sofia", "Anthony", "Scarlett", "Caleb", "Victoria", "Logan", "Madison", "Isaac", "Eleanor", "Lucas", "Hannah", "Owen", "Addison", "Dylan", "Zoe", "Jack", "Penelope", "Luke", "Layla", "Jeremiah", "Natalie", "Isaiah", "Audrey", "Carter", "Leah", "Josiah", "Savannah", "Julian", "Brooklyn", "Wyatt", "Stella", "Hunter", "Claire", "Levi", "Skylar", "Christian", "Maya", "Eli", "Paisley", "Lincoln", "Anna", "Jordan", "Caroline", "Charles", "Eliana", "Thomas", "Ruby", "Aaron", "Aria", "Connor", "Aurora", "Cameron", "Naomi", "Adrian", "Valentina", "Landon", "Alexa", "Gavin", "Lydia", "Evan", "Piper", "Sebastian", "Ariana", "Cooper", "Sadie"}
	users := make([]utils.User, len(names))
	for i := 0; i < len(names); i++ {
		a, _ := rand.Int(rand.Reader, bn256.Order)
		b, _ := rand.Int(rand.Reader, bn256.Order)
		A := new(bn256.G1).ScalarBaseMult(a)
		B := new(bn256.G1).ScalarBaseMult(b)
		para := []interface{}{"Register", names[i], contract.EmailPK{utils.G1ToPoint(A), utils.G1ToPoint(B)}}
		privatekey := utils.GetENV("PRIVATE_KEY_1")
		//fmt.Println(privatekey, i%10+1)
		_ = utils.Transact(client, privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
		users[i] = utils.User{names[i], a, b, A, B, privatekey, nil}
	}
	//users generate their BIP32 child keys
	utils.InitBIP32Wallet(client, users)

	Bob := users[0]
	Alice := users[1]
	//Bob generates Alice's Stealth address after downloading Alice's public keys
	fmt.Println("=============================test one-to-one mailing=====================")
	m, _ := rand.Int(rand.Reader, bn256.Order)
	key := new(bn256.G1).ScalarBaseMult(m)
	msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00. \nBest,\nBob")
	recs := []string{names[8], names[9]} //user names to confuse others
	utils.MailTo(client, ctc, Bob, key, msg, Alice, recs)
	//Alice downloads the encrypted email
	utils.ReadMail(ctc, Alice)

	//Bob is a group administrator
	n := len(users)
	grpId := "group1"
	brdPks, brdPrivs := broadcast.Setup(n, grpId)
	Bobindex := 0
	Bob = users[Bobindex]
	//Bob.Groups = map[string]utils.Groups{grpId: {brdPks, brdPrivs[Bobindex], nil}}
	fmt.Println("=============================upload broadcast public keys=====================")
	//Bob sends brdPrivs to each group member via one-to-one mailing (Here, sends secretkeys to the domain)
	utils.RegGroup(client, ctc, Bob, brdPks, brdPrivs, users)

	////Charlie is a domain manager, Charlie generates \prod_j∈S g_{n+1-j} for the domain
	fmt.Println("=============================build domain public keys=====================")
	size := n / 2
	S := make([]uint32, size)
	for i := 0; i < size; i++ {
		S[i] = uint32(i) + 1
	}
	domainId := "computer@" + grpId
	Charlie := users[2]
	fmt.Println(domainId)
	utils.RegDomain(client, ctc, Charlie, domainId, S)
	//todo multiple domains

	// Emily as a member, broadcast a email
	fmt.Println("=============================broadcast email 1=====================")
	Emily := users[3]
	//todo download grpId and domainId
	pArrEmily, qArrEmily, vEmily, _ := ctc.GetBrdPKs(&bind.CallOpts{}, grpId)
	SEmily, _ := ctc.GetS(&bind.CallOpts{}, domainId)
	//fmt.Println(SEmily)
	brdPksEmily := broadcast.PKs{utils.PointsToG1(pArrEmily), utils.PointsToG2(qArrEmily), *utils.PointToG1(vEmily), grpId}
	brdPrivEmily := utils.DownloadAndResolvePriv(ctc, Emily, grpId)
	Emily.Groups = map[string]utils.Group{grpId: {brdPksEmily, brdPrivEmily, map[string][]uint32{domainId: SEmily}}}
	msgEmily := []byte("Dear Staff, we are going to have a meeting at Jun 30, 2024 09:00 at the gym. ---Emily")
	utils.BroadcastTo(client, ctc, Emily, msgEmily, domainId)

	// Alexander as a member, broadcast a email
	fmt.Println("=============================broadcast email 2=====================")
	Alexander := users[4]
	pArrAlexander, qArrAlexander, vAlexander, _ := ctc.GetBrdPKs(&bind.CallOpts{}, grpId)
	SAlexander, _ := ctc.GetS(&bind.CallOpts{}, domainId)
	//fmt.Println(SAlexander)
	brdPksAlexander := broadcast.PKs{utils.PointsToG1(pArrAlexander), utils.PointsToG2(qArrAlexander), *utils.PointToG1(vAlexander), grpId}
	brdPrivAlexander := utils.DownloadAndResolvePriv(ctc, Alexander, grpId)
	Alexander.Groups = map[string]utils.Group{grpId: {brdPksAlexander, brdPrivAlexander, map[string][]uint32{domainId: SAlexander}}}
	//map[string]utils.Groups{grpId:{brdPksAlice, brdPrivAlice, map[string][]uint32{domainId: SAlice}}}
	msgAlexander := []byte("Dear Staff, I am Alexander. ---Alexander")
	utils.BroadcastTo(client, ctc, Alexander, msgAlexander, domainId)

	//Alice is a domain and ties to read the email
	fmt.Println("=============================read broadcasted email=====================")
	Alice = users[1]
	groupIds, _ := ctc.GetMyGroups(&bind.CallOpts{}, Alice.Psid)

	for i := 0; i < len(groupIds); i++ {
		SAlice, _ := ctc.GetS(&bind.CallOpts{}, domainId)
		pArrAlice, qArrAlice, vAlice, _ := ctc.GetBrdPKs(&bind.CallOpts{}, groupIds[i])
		brdPksAlice := broadcast.PKs{utils.PointsToG1(pArrAlice), utils.PointsToG2(qArrAlice), *utils.PointToG1(vAlice), groupIds[i]}
		brdPrivAlice := utils.DownloadAndResolvePriv(ctc, Alice, groupIds[i])
		//fmt.Println(i, "brdPrivAlice", brdPrivAlice.Di.String()[:30])
		Alice.Groups = map[string]utils.Group{grpId: {brdPksAlice, brdPrivAlice, map[string][]uint32{domainId: SAlice}}}
	}
	utils.ReadBrdMail(ctc, Alice, domainId)
}

//pairingRes, _ := ctc.GetPairingRes(&bind.CallOpts{})
//fmt.Printf("GetPairingRes: %v\n", pairingRes)
