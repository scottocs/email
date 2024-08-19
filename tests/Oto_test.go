package tests

import (
	"email/utils"
	"fmt"
	"testing"
)

func TestOto(t *testing.T) {
	users, client, ctc := utils.DeployAndInitWallet()
	// names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy", "Isabella", "Jacob", "Ava", "Matthew", "Mia", "Daniel", "Abigail", "Ethan", "Harper", "Max", "Amelia", "Ryan", "Evelyn", "Nathan", "Elizabeth", "Samuel", "Charlotte", "Christopher", "Grace", "Jonathan", "Lily", "Gabriel", "Ella", "Andrew", "Avery", "Joshua", "Sofia", "Anthony", "Scarlett", "Caleb", "Victoria", "Logan", "Madison", "Isaac", "Eleanor", "Lucas", "Hannah", "Owen", "Addison", "Dylan", "Zoe", "Jack", "Penelope", "Luke", "Layla", "Jeremiah", "Natalie", "Isaiah", "Audrey", "Carter", "Leah", "Josiah", "Savannah", "Julian", "Brooklyn", "Wyatt", "Stella", "Hunter", "Claire", "Levi", "Skylar", "Christian", "Maya", "Eli", "Paisley", "Lincoln", "Anna", "Jordan", "Caroline", "Charles", "Eliana", "Thomas", "Ruby", "Aaron", "Aria", "Connor", "Aurora", "Cameron", "Naomi", "Adrian", "Valentina", "Landon", "Alexa", "Gavin", "Lydia", "Evan", "Piper", "Sebastian", "Ariana", "Cooper", "Sadie"}
	indexAlice := 1
	Alice := utils.ResolveUser(ctc, users[indexAlice].Psid, users[indexAlice].Aa, users[indexAlice].Bb, users[indexAlice].Privatekey, users[indexAlice].Addr)
	for i := 0; i < 5; i++ {
		sender := users[i]
		fmt.Println("=============================one-to-one mailing=====================", i)
		msg := []byte("Alice, I am inviting you to have a dinner at Jun 29, 2024 18:00. ------" + sender.Psid)
		recs := []string{Alice.Psid} //user names to confuse others
		for j := 0; j < 8; j++ {
			recs = append(recs, names[i+2+j])
		}
		utils.Oto(client, ctc, sender, msg, Alice, recs)
	}
	utils.ReadMail(ctc, Alice)

}
