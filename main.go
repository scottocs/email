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

	//Bob is a domain administrator
	n := len(users)
	domainId := "domain1"
	brdPks, brdPrivs := broadcast.Setup(n, domainId)
	Bobindex := 0
	Bob = users[Bobindex]
	//Bob.Domains = map[string]utils.Domains{domainId: {brdPks, brdPrivs[Bobindex], nil}}
	fmt.Println("=============================upload broadcast public keys=====================")
	//Bob sends brdPrivs to each domain member via one-to-one mailing (Here, sends secretkeys to the cluster)
	utils.RegDomain(client, ctc, Bob, brdPks, brdPrivs, users)

	////Charlie is a cluster manager, Charlie generates \prod_j∈S g_{n+1-j} for the cluster
	fmt.Println("=============================build cluster public keys=====================")
	size := n / 2
	S := make([]uint32, size)
	for i := 0; i < size; i++ {
		S[i] = uint32(i) + 1
	}
	clusterId := "computer@" + domainId
	Charlie := users[2]
	fmt.Println(clusterId)
	utils.RegCluster(client, ctc, Charlie, clusterId, S)
	//todo multiple clusters

	// Emily as a member, broadcast a email
	fmt.Println("=============================broadcast email 1=====================")
	Emily := users[3]
	//todo download domainId and clusterId
	pArrEmily, qArrEmily, vEmily, _ := ctc.GetBrdPKs(&bind.CallOpts{}, domainId)
	SEmily, _ := ctc.GetS(&bind.CallOpts{}, clusterId)
	//fmt.Println(SEmily)
	brdPksEmily := broadcast.PKs{utils.PointsToG1(pArrEmily), utils.PointsToG2(qArrEmily), *utils.PointToG1(vEmily), domainId}
	brdPrivEmily := utils.DownloadAndResolvePriv(ctc, Emily, domainId)
	Emily.Domains = map[string]utils.Domain{domainId: {brdPksEmily, brdPrivEmily, map[string][]uint32{clusterId: SEmily}}}
	msgEmily := []byte("Dear Staff, we are going to have a meeting at Jun 30, 2024 09:00 at the gym. ---Emily")
	utils.BroadcastTo(client, ctc, Emily, msgEmily, clusterId)

	// Alexander as a member, broadcast a email
	fmt.Println("=============================broadcast email 2=====================")
	Alexander := users[4]
	pArrAlexander, qArrAlexander, vAlexander, _ := ctc.GetBrdPKs(&bind.CallOpts{}, domainId)
	SAlexander, _ := ctc.GetS(&bind.CallOpts{}, clusterId)
	//fmt.Println(SAlexander)
	brdPksAlexander := broadcast.PKs{utils.PointsToG1(pArrAlexander), utils.PointsToG2(qArrAlexander), *utils.PointToG1(vAlexander), domainId}
	brdPrivAlexander := utils.DownloadAndResolvePriv(ctc, Alexander, domainId)
	Alexander.Domains = map[string]utils.Domain{domainId: {brdPksAlexander, brdPrivAlexander, map[string][]uint32{clusterId: SAlexander}}}
	//map[string]utils.Domains{domainId:{brdPksAlice, brdPrivAlice, map[string][]uint32{clusterId: SAlice}}}
	msgAlexander := []byte("Dear Staff, I am Alexander. ---Alexander")
	utils.BroadcastTo(client, ctc, Alexander, msgAlexander, clusterId)

	//Alice is a cluster and ties to read the email
	fmt.Println("=============================read broadcasted email=====================")
	Alice = users[1]
	domainIds, _ := ctc.GetMyDomains(&bind.CallOpts{}, Alice.Psid)

	for i := 0; i < len(domainIds); i++ {
		SAlice, _ := ctc.GetS(&bind.CallOpts{}, clusterId)
		pArrAlice, qArrAlice, vAlice, _ := ctc.GetBrdPKs(&bind.CallOpts{}, domainIds[i])
		brdPksAlice := broadcast.PKs{utils.PointsToG1(pArrAlice), utils.PointsToG2(qArrAlice), *utils.PointToG1(vAlice), domainIds[i]}
		brdPrivAlice := utils.DownloadAndResolvePriv(ctc, Alice, domainIds[i])
		//fmt.Println(i, "brdPrivAlice", brdPrivAlice.Di.String()[:30])
		Alice.Domains = map[string]utils.Domain{domainId: {brdPksAlice, brdPrivAlice, map[string][]uint32{clusterId: SAlice}}}
	}
	utils.ReadBrdMail(ctc, Alice, clusterId)
}

//pairingRes, _ := ctc.GetPairingRes(&bind.CallOpts{})
//fmt.Printf("GetPairingRes: %v\n", pairingRes)
