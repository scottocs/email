package tests

import (
	"email/crypto/broadcast"
	"email/utils"
	"fmt"
	"strings"
	"testing"
)

func TestBcstLinkableCluster(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	//Bob is a domain administrator
	n := len(users)
	domainId := "google"
	brdPks, brdPrivs := broadcast.Setup(n, domainId)
	Bobindex := 0
	Bob := users[Bobindex]
	//Bob.Domains = map[string]utils.Domains{domainId: {brdPks, brdPrivs[Bobindex], nil}}
	fmt.Println("\n\n=============================register domain=====================")
	//Bob sends brdPrivs to each domain member via one-to-one mailing (Here, sends secretkeys to the cluster)
	utils.RegDomain(client, ctc, Bob, brdPks, brdPrivs, users)

	////Charlie is a cluster manager, Charlie generates \prod_j∈S g_{n+1-j} for the cluster
	fmt.Println("=============================register cluster=====================")
	size := n
	NArr := make([]uint32, size)
	for i := 0; i < size; i++ {
		NArr[i] = uint32(i) + 1
	}
	ClS := NArr[0:n]
	fmt.Printf("domain size: %d, cluster size:%d\n", len(users), len(NArr))
	clusterId := "android@" + domainId
	// cluster should be created by the domain administrator

	utils.RegCluster(client, ctc, Bob, clusterId, brdPks, NArr, ClS)
	//todo multiple clusters

	// Emily as a member, broadcast a email
	fmt.Println("=============================broadcast email 1=====================")
	indexEmily := 3
	Emily := utils.ResolveUser(ctc, users[indexEmily].Psid, users[indexEmily].Aa, users[indexEmily].Bb, users[indexEmily].Privatekey, users[indexEmily].Addr)
	msgEmily := []byte("Dear Staff, we are going to have a meeting at Jun 30, 2024 09:00 at the gym. ---Emily")
	utils.BroadcastTo(client, ctc, Emily, msgEmily, clusterId)

	indexAlexander := 4
	Alexander := utils.ResolveUser(ctc, users[indexAlexander].Psid, users[indexAlexander].Aa, users[indexAlexander].Bb, users[indexAlexander].Privatekey, users[indexAlexander].Addr)
	for i := 0; i < 3; i++ {
		// Alexander as a member, broadcast a email
		fmt.Println("=============================broadcast email 2=====================", i)
		msgAlexander := []byte("Dear Staff, I am Alexander. ---Alexander")
		utils.BroadcastTo(client, ctc, Alexander, msgAlexander, clusterId)
	}

	//Alice is a cluster and ties to read the email
	fmt.Println("=============================read broadcasted email=====================")
	indexAlice := 1
	Alice := utils.ResolveUser(ctc, "Alice", users[indexAlice].Aa, users[indexAlice].Bb, users[indexAlice].Privatekey, users[indexAlice].Addr)
	utils.ReadBrdMail(ctc, Alice)

	fmt.Println("=============================withdraw deposited money=====================")
	utils.Reward(client, ctc, Alice, domainId)
	//utils.Reward(client, ctc, Alice, domainId)
}

func TestBcstOneshotCluster(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	psids := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy", "Isabella", "Jacob", "Ava", "Matthew", "Mia", "Daniel", "Abigail", "Ethan", "Harper", "Max", "Amelia", "Ryan", "Evelyn", "Nathan", "Elizabeth", "Samuel", "Charlotte", "Christopher", "Grace", "Jonathan", "Lily", "Gabriel", "Ella", "Andrew", "Avery", "Joshua", "Sofia", "Anthony", "Scarlett", "Caleb", "Victoria", "Logan", "Madison", "Isaac", "Eleanor", "Lucas", "Hannah", "Owen", "Addison", "Dylan", "Zoe", "Jack", "Penelope", "Luke", "Layla", "Jeremiah", "Natalie", "Isaiah", "Audrey", "Carter", "Leah", "Josiah", "Savannah", "Julian", "Brooklyn", "Wyatt", "Stella", "Hunter", "Claire", "Levi", "Skylar", "Christian", "Maya", "Eli", "Paisley", "Lincoln", "Anna", "Jordan", "Caroline", "Charles", "Eliana", "Thomas", "Ruby", "Aaron", "Aria", "Connor", "Aurora", "Cameron", "Naomi", "Adrian", "Valentina", "Landon", "Alexa", "Gavin", "Lydia", "Evan", "Piper", "Sebastian", "Ariana", "Cooper", "Sadie"}

	fmt.Println("\n\n=============================Bob creates temperory domain and cluster =====================")
	createdUsers, createdClsId, ClS := utils.CreateTempCluster(client, ctc, users[0], psids)
	tmpUser := utils.ResolveTmpUser(ctc, users[0].Psid, users[0].Aa, users[0].Bb, users[0].Addr, 0)
	fmt.Printf("domain size: %d, cluster size:%d\n", len(createdUsers), len(ClS))
	for i := 0; i < 3; i++ {
		fmt.Println("=============================Bob broadcast =====================", i)
		msgCreated := []byte("Dear there, run -------by " + tmpUser[0].Psid)
		utils.BroadcastTo(client, ctc, tmpUser[0], msgCreated, createdClsId)
	}

	//Alice is a cluster and ties to read the email
	fmt.Println("=============================createUser read broadcasted email=====================")
	num := len(createdUsers)
	if num > 10 {
		num = 10
	}
	for j := 0; j < num; j++ {
		user := utils.ResolveUser(ctc, psids[j], users[j].Aa, users[j].Bb, users[j].Privatekey, users[j].Addr)
		tmpUsers := utils.ResolveTmpUser(ctc, user.Psid, user.Aa, user.Bb, user.Addr, j)
		for i := 0; i < len(tmpUsers); i++ {
			utils.ReadBrdMail(ctc, tmpUsers[i])
		}
	}

	fmt.Println("=============================withdraw deposited money=====================")
	_2dmId := strings.Split(createdClsId, "@")
	utils.Reward(client, ctc, tmpUser[0], _2dmId[1])
}
