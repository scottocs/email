package tests

import (
	"email/crypto/broadcast"
	"email/utils"
	"fmt"
	"testing"
)

func TestBcstOneshotCluster(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	psids := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	//Bob := users[0]
	fmt.Println("=============================Bob creates temperory domain and broadcast=====================")
	createdUsers, createdClsId := utils.CreateTempCluster(client, ctc, users[0], psids)
	msgCreated := []byte("Dear there, run. -------by " + createdUsers[0].Psid)
	utils.BroadcastTo(client, ctc, createdUsers[0], msgCreated, createdClsId)

	//Alice is a cluster and ties to read the email
	fmt.Println("=============================createUser read broadcasted email=====================")
	for j := 0; j < len(createdUsers); j++ {
		User := utils.ResolveUser(ctc, psids[j], users[j].Aa, users[j].Bb, users[j].Privatekey, users[j].Addr)
		TmpUsers := utils.ResolveTmpUser(ctc, User.Psid, User.Aa, User.Bb, User.Addr)
		for i := 0; i < len(TmpUsers); i++ {
			utils.ReadBrdMail(ctc, TmpUsers[i])
		}
	}
}

func TestBcstLinkableCluster(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	//Bob is a domain administrator
	n := len(users)
	domainId := "google"
	brdPks, brdPrivs := broadcast.Setup(n, domainId)
	Bobindex := 0
	Bob := users[Bobindex]
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
	//fmt.Println(clusterId)
	utils.RegCluster(client, ctc, Charlie, clusterId, S)
	//todo multiple clusters

	// Emily as a member, broadcast a email
	fmt.Println("=============================broadcast email 1=====================")
	indexEmily := 3
	Emily := utils.ResolveUser(ctc, "Emily", users[indexEmily].Aa, users[indexEmily].Bb, users[indexEmily].Privatekey, users[indexEmily].Addr)
	msgEmily := []byte("Dear Staff, we are going to have a meeting at Jun 30, 2024 09:00 at the gym. ---Emily")
	utils.BroadcastTo(client, ctc, Emily, msgEmily, clusterId)

	// Alexander as a member, broadcast a email
	fmt.Println("=============================broadcast email 2=====================")
	indexAlexander := 4
	Alexander := utils.ResolveUser(ctc, "Alexander", users[indexAlexander].Aa, users[indexAlexander].Bb, users[indexAlexander].Privatekey, users[indexAlexander].Addr)
	msgAlexander := []byte("Dear Staff, I am Alexander. ---Alexander")
	utils.BroadcastTo(client, ctc, Alexander, msgAlexander, clusterId)

	//Alice is a cluster and ties to read the email
	fmt.Println("=============================read broadcasted email=====================")
	indexAlice := 1
	Alice := utils.ResolveUser(ctc, "Alice", users[indexAlice].Aa, users[indexAlice].Bb, users[indexAlice].Privatekey, users[indexAlice].Addr)
	utils.ReadBrdMail(ctc, Alice)

}

//pairingRes, _ := ctc.GetPairingRes(&bind.CallOpts{})
//fmt.Printf("GetPairingRes: %v\n", pairingRes)
