package tests

import (
	"email/utils"
	"fmt"
	"testing"
)

func TestOto(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	indexAlice := 1
	Alice := utils.ResolveUser(ctc, users[indexAlice].Psid, users[indexAlice].Aa, users[indexAlice].Bb, users[indexAlice].Privatekey, users[indexAlice].Addr)
	for i := 0; i < 5; i++ {
		sender := users[i]
		fmt.Println("=============================one-to-one mailing=====================", i)
		msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00. ------" + sender.Psid)
		recs := []string{Alice.Psid, names[i+2], names[i+3]} //user names to confuse others
		utils.MailTo(client, ctc, sender, msg, Alice, recs)
	}
	utils.ReadMail(ctc, Alice)

}
