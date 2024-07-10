package tests

import (
	"crypto/rand"
	"email/utils"
	"fmt"
	"github.com/fentec-project/bn256"
	"testing"
)

func TestOto(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	Bob := users[0]
	//Alice := users[1]
	indexAlice := 1
	Alice := utils.ResolveUser(ctc, "Alice", users[indexAlice].Aa, users[indexAlice].Bb, users[indexAlice].Privatekey, users[indexAlice].Addr)
	//Bob generates Alice's Stealth address after downloading Alice's public keys
	fmt.Println("=============================test one-to-one mailing=====================")
	m, _ := rand.Int(rand.Reader, bn256.Order)
	key := new(bn256.G1).ScalarBaseMult(m)
	msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00. ------Bob")
	recs := []string{names[8], names[9]} //user names to confuse others
	utils.MailTo(client, ctc, Bob, key, msg, Alice, recs)
	utils.ReadMail(ctc, Alice)

}

//pairingRes, _ := ctc.GetPairingRes(&bind.CallOpts{})
//fmt.Printf("GetPairingRes: %v\n", pairingRes)
