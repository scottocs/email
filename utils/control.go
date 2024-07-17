package utils

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"email/compile/contract"
	"email/crypto/aes"
	"email/crypto/broadcast"
	"email/crypto/stealth"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fentec-project/bn256"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func DeployAndInitWallet() ([]User, *ethclient.Client, *contract.Contract) {
	ether01 := big.NewInt(100000)
	//ether := big.NewInt(1000000000000000000)
	//ether10 := big.NewInt(1).Mul(ether, big.NewInt(10))
	contract_name := "Email"
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	//Deploy
	deployTX := Transact(client, GetENV("PRIVATE_KEY_1"), big.NewInt(0), nil, nil)
	address, _ := Deploy(client, contract_name, deployTX.(*bind.TransactOpts))
	ctc, err := contract.NewContract(common.HexToAddress(address.Hex()), client)

	ioutil.WriteFile("../compile/contract/addr.txt", []byte(address.String()), 0644)

	fmt.Println("=============================register psids and public keys=====================")
	//Users register their public keys (A B)
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	//names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy", "Isabella", "Jacob", "Ava", "Matthew", "Mia", "Daniel", "Abigail", "Ethan", "Harper", "Alexander", "Amelia", "Ryan", "Evelyn", "Nathan", "Elizabeth", "Samuel", "Charlotte", "Christopher", "Grace", "Jonathan", "Lily", "Gabriel", "Ella", "Andrew", "Avery", "Joshua", "Sofia", "Anthony", "Scarlett", "Caleb", "Victoria", "Logan", "Madison", "Isaac", "Eleanor", "Lucas", "Hannah", "Owen", "Addison", "Dylan", "Zoe", "Jack", "Penelope", "Luke", "Layla", "Jeremiah", "Natalie", "Isaiah", "Audrey", "Carter", "Leah", "Josiah", "Savannah", "Julian", "Brooklyn", "Wyatt", "Stella", "Hunter", "Claire", "Levi", "Skylar", "Christian", "Maya", "Eli", "Paisley", "Lincoln", "Anna", "Jordan", "Caroline", "Charles", "Eliana", "Thomas", "Ruby", "Aaron", "Aria", "Connor", "Aurora", "Cameron", "Naomi", "Adrian", "Valentina", "Landon", "Alexa", "Gavin", "Lydia", "Evan", "Piper", "Sebastian", "Ariana", "Cooper", "Sadie"}
	users := make([]User, len(names))
	for i := 0; i < len(names); i++ {
		a, _ := rand.Int(rand.Reader, bn256.Order)
		b, _ := rand.Int(rand.Reader, bn256.Order)
		A := new(bn256.G1).ScalarBaseMult(a)
		B := new(bn256.G1).ScalarBaseMult(b)
		privatekey := GetENV("PRIVATE_KEY_" + strconv.Itoa(i+1))
		key, _ := crypto.HexToECDSA(ReverseString(privatekey))
		publicKey := key.Public()
		publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
		addr := crypto.PubkeyToAddress(*publicKeyECDSA)
		para := []interface{}{"Register", names[i], contract.EmailPK{G1ToPoint(A), G1ToPoint(B), ether01, addr, make([]contract.EmailG1Point, 0)}}
		_ = Transact(client, privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
		var domains = make(map[string]Domain)
		users[i] = User{names[i], a, b, A, B, privatekey, addr.String(), domains}
		//users[i] = ResolveUser(ctc, "Emily", a, b, "", "", i)
	}
	//users generate their BIP32 child keys
	InitBIP32Wallet(client, users)
	return users, client, ctc
}
func MailTo(client *ethclient.Client, ctc *contract.Contract, sender User, key *bn256.G1, msg []byte, to User, recs []string) string {
	pkRes, _ := ctc.GetPK(&bind.CallOpts{}, to.Psid)
	sa := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)})
	r, _ := rand.Int(rand.Reader, bn256.Order)
	c1 := new(bn256.G1).ScalarBaseMult(r) // c1 = r * G
	c2 := new(bn256.G1).Add(new(bn256.G1).ScalarMult(sa.S, r), key)
	ct, _ := aes.Encrypt(msg, key.Marshal()[:32])
	cid := IPFSUpload(ct)
	mail := contract.EmailMail{contract.EmailStealthPub{G1ToPoint(sa.R), G1ToPoint(sa.S)}, contract.EmailElGamalCT{G1ToPoint(c1), G1ToPoint(c2)}}
	para := []interface{}{"MailTo", mail, cid, append(recs, to.Psid)}
	ether := big.NewInt(1000000000000000000)
	ether100 := big.NewInt(1).Mul(ether, big.NewInt(100))
	//fmt.Println(sender)
	_ = Transact(client, sender.Privatekey, ether100, ctc, para).(*types.Receipt)

	return cid
}

func ReadMail(ctc *contract.Contract, my User) {
	currentTime := time.Now()
	timestamp := currentTime.Unix()
	dayTS := timestamp - (timestamp % 86400)

	cids, dayMails, _ := ctc.GetDailyMail(&bind.CallOpts{}, my.Psid, uint64(dayTS))
	//fmt.Println(uint64(dayTS), dayMails)
	for i := 0; i < len(cids); i++ {
		sp := stealth.ResolvePriv(stealth.SecretKey{my.Aa, my.Bb},
			stealth.StealthPub{PointToG1(dayMails[i].Pub.R), PointToG1(dayMails[i].Pub.S)})
		cid2Flag := strings.Split(cids[i], "||")
		//fmt.Println(cid2Flag)
		GetIPFSClient().Get(cid2Flag[0], "../users/"+my.Psid+"/")
		file, _ := os.Open("../users/" + my.Psid + "/" + cid2Flag[0])
		content, _ := io.ReadAll(file)
		decRes := string(content)
		//if cid2Flag[1] == "0" {
		c1pNeg := new(bn256.G1).Neg(PointToG1(dayMails[i].Ct.C1))
		c2p := PointToG1(dayMails[i].Ct.C2)
		keyp := new(bn256.G1).Add(c2p, new(bn256.G1).ScalarMult(c1pNeg, sp))
		decRes, _ = aes.Decrypt(decRes, keyp.Marshal()[:32])
		//}
		fmt.Println("Email content (read):", decRes)
	}

}
func RegCluster(client *ethclient.Client, ctc *contract.Contract, from User, clsId string, S []uint32) {
	para := []interface{}{"RegCluster", clsId, S}
	_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
}

func RegDomain(client *ethclient.Client, ctc *contract.Contract, from User, cpk broadcast.PKs, brdPrivs []broadcast.SK, to []User) {

	cts := make([]contract.EmailElGamalCT, len(to))
	names := make([]string, len(brdPrivs))
	for i := 0; i < len(to); i++ {
		pkRes, _ := ctc.GetPK(&bind.CallOpts{}, to[i].Psid)
		//sa := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)})
		// todo whether a stealth address is needed?

		// ElGamal encrypted
		r, _ := rand.Int(rand.Reader, bn256.Order)
		cts[i].C1 = G1ToPoint(new(bn256.G1).ScalarBaseMult(r))
		cts[i].C2 = G1ToPoint(new(bn256.G1).Add(new(bn256.G1).ScalarMult(PointToG1(pkRes.A), r), &brdPrivs[i+1].Di))
		names[i] = to[i].Psid
		//fmt.Println("------------", names[i], cts[i].C1, len(cts), len(names))
	}
	//fmt.Println(len(cpk.PArr), len(cpk.QArr), len(to)) //2n+1,n+1,n
	para := []interface{}{"RegDomain", cpk.DmId, G1ArrToPoints(cpk.PArr), G2ArrToPoints(cpk.QArr),
		G1ToPoint(&cpk.V), cts, names}
	_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
}

//	func DownloadPKs(ctc *contract.Contract, clsId string) broadcast.SK {
//		dmId := strings.Split(clsId, "@")[1]
//		pArr, qArr, v, _ := ctc.GetBrdPKs(&bind.CallOpts{}, dmId)
//		S, _ := ctc.GetS(&bind.CallOpts{}, clsId)
//		//fmt.Println(S)
//		brdPks := broadcast.PKs{PointsToG1(pArr), PointsToG2(qArr), *PointToG1(v), dmId}
//	}

func ResolveTmpUser(ctc *contract.Contract, psid string, Aa *big.Int, Bb *big.Int, wallet string) []User {
	tmpPsids, _ := ctc.GetTmpPsid(&bind.CallOpts{}, psid)
	users := make([]User, 0)
	for i := 0; i < len(tmpPsids); i++ {
		DomainIds, _ := ctc.GetMyDomains(&bind.CallOpts{}, tmpPsids[i])
		var domains = make(map[string]Domain)
		pkRes, _ := ctc.GetPK(&bind.CallOpts{}, tmpPsids[i])
		S1 := PointToG1(pkRes.A)
		S2 := PointToG1(pkRes.B)
		R1 := PointToG1(pkRes.Extra[0])
		R2 := PointToG1(pkRes.Extra[1])
		a := stealth.ResolvePriv(stealth.SecretKey{Aa, Bb}, stealth.StealthPub{R1, S1})
		b := stealth.ResolvePriv(stealth.SecretKey{Aa, Bb}, stealth.StealthPub{R2, S2})
		user := User{tmpPsids[i], a, b, S1, S2,
			"", wallet, domains}
		for k := 0; k < len(DomainIds); k++ {
			dmId := DomainIds[k].DmId
			index := DomainIds[k].Index
			index, ct, _ := ctc.GetBrdEncPrivs(&bind.CallOpts{}, dmId, user.Psid)
			c1pNeg := new(bn256.G1).Neg(PointToG1(ct.C1))
			myBrdPriv := new(bn256.G1).Add(PointToG1(ct.C2), new(bn256.G1).ScalarMult(c1pNeg, user.Aa))
			//fmt.Println("Aa hash", stealth.Hash2Int(new(bn256.G1).ScalarMult(sa1R, my.Bb).String()))
			//fmt.Println("222222222", dmId, myBrdPriv.String()[:30])
			user.Domains = make(map[string]Domain)
			pArr, qArr, v, _ := ctc.GetBrdPKs(&bind.CallOpts{}, dmId)
			brdPks := broadcast.PKs{PointsToG1(pArr), PointsToG2(qArr), *PointToG1(v), dmId}
			var clsIds = map[string][]uint32{}
			clsIdsDL, _ := ctc.GetMyClusters(&bind.CallOpts{}, dmId)
			for j := 0; j < len(clsIdsDL); j++ {
				clsId := clsIdsDL[j]
				S, _ := ctc.GetS(&bind.CallOpts{}, clsId)
				clsIds[clsId] = S
			}
			user.Domains[dmId] = Domain{brdPks, broadcast.SK{int(index.Int64()) + 1, *myBrdPriv}, clsIds}
		}
		//fmt.Println("user", user)
		users = append(users, user)
	}
	return users
}

func ResolveUser(ctc *contract.Contract, psid string, Aa *big.Int, Bb *big.Int, priStr string, addr string) User {
	DomainIds, _ := ctc.GetMyDomains(&bind.CallOpts{}, psid)
	A := new(bn256.G1).ScalarBaseMult(Aa)
	B := new(bn256.G1).ScalarBaseMult(Bb)
	var domains = make(map[string]Domain)
	user := User{psid, Aa, Bb, A, B, priStr, addr, domains}
	for i := 0; i < len(DomainIds); i++ {
		dmId := DomainIds[i].DmId
		index := DomainIds[i].Index
		pArr, qArr, v, _ := ctc.GetBrdPKs(&bind.CallOpts{}, dmId)
		brdPks := broadcast.PKs{PointsToG1(pArr), PointsToG2(qArr), *PointToG1(v), dmId}

		_, c, _ := ctc.GetBrdEncPrivs(&bind.CallOpts{}, dmId, psid)
		c1pNeg := new(bn256.G1).Neg(PointToG1(c.C1))
		myBrdPriv := new(bn256.G1).Add(PointToG1(c.C2), new(bn256.G1).ScalarMult(c1pNeg, Aa))
		var clsIds = map[string][]uint32{}

		clsIdsDL, _ := ctc.GetMyClusters(&bind.CallOpts{}, dmId)

		for j := 0; j < len(clsIdsDL); j++ {
			clsId := clsIdsDL[j]
			S, _ := ctc.GetS(&bind.CallOpts{}, clsId)
			clsIds[clsId] = S
		}
		user.Domains[dmId] = Domain{brdPks, broadcast.SK{int(index.Int64()), *myBrdPriv}, clsIds}
	}
	//fmt.Println("user", user)
	return user
}

func BroadcastTo(client *ethclient.Client, ctc *contract.Contract, sender User, msg []byte, clusterId string) string {
	// todo download cpk
	dmId := strings.Split(clusterId, "@")[1]
	brdPKs := sender.Domains[dmId].PKs
	hdr, beK := brdPKs.Encrypt(sender.Domains[dmId].Clusters[clusterId])
	//fmt.Println("beK", clusterId, sender.Domains[dmId].Clusters[clusterId], clusterId, beK.String()[:30], "hdr", brdPKs.V.String()[:30])
	//fmt.Println("beK.String()", beK.String()[:30])
	ct, _ := aes.Encrypt(msg, beK.Marshal()[:32])
	fmt.Println("encrypted email to broadcast:", ct)
	// Bob uploads encrypted email content to IPFS
	cid, _ := GetIPFSClient().Add(strings.NewReader(ct))
	fmt.Println("broadcast mail IPFS link:", cid)
	x, _ := rand.Int(rand.Reader, bn256.Order)
	xp, _ := rand.Int(rand.Reader, bn256.Order)
	//x = big.NewInt(1)
	senderIndex := sender.Domains[dmId].SK.I
	clusterRecivers := contract.EmailBcstHeader{G1ToPoint(hdr.C0), G1ToPoint(hdr.C1), G2ToPoint(hdr.C0p)}
	ptr := sender.Domains[dmId].SK.Di
	C := new(bn256.G1).ScalarMult(&brdPKs.V, x)
	Cp := new(bn256.G1).ScalarMult(&brdPKs.V, xp)
	hash := sha256.Sum256([]byte(C.String() + Cp.String()))
	c := new(big.Int).SetBytes(hash[:])
	hatc := new(big.Int).Sub(xp, new(big.Int).Mul(x, c))
	hatc = new(big.Int).Mod(hatc, bn256.Order)
	proof := contract.EmailDomainProof{G1ToPoint(new(bn256.G1).ScalarMult(&ptr, x)),
		G2ToPoint(&brdPKs.QArr[senderIndex]), G1ToPoint(C), G1ToPoint(Cp), c, hatc,
	}
	//e(skipows,g2)= e(pki,vpows)
	// schnorr sigma protocol verification
	//fmt.Println("point", new(bn256.G1).ScalarMult(&brdPKs.V, hatc).String())
	para := []interface{}{"BcstTo", clusterRecivers, clusterId, proof, cid}
	ether := big.NewInt(1000000000000000000)
	ether100 := big.NewInt(1).Mul(ether, big.NewInt(100))
	_ = Transact(client, sender.Privatekey, ether100, ctc, para).(*types.Receipt)
	//SCpoint, _ := ctc.GetPoint(&bind.CallOpts{})
	//fmt.Println("sc point", PointToG1(SCpoint))
	return cid
}

func ReadBrdMail(ctc *contract.Contract, created User) {
	DomainIds, _ := ctc.GetMyDomains(&bind.CallOpts{}, created.Psid)
	for i := 0; i < len(DomainIds); i++ {
		dmId := DomainIds[i].DmId
		clsIdsDL, _ := ctc.GetMyClusters(&bind.CallOpts{}, dmId)
		for j := 0; j < len(clsIdsDL); j++ {
			clusterId := clsIdsDL[j]
			//dmId := strings.Split(clusterId, "@")[1]
			currentTime := time.Now()
			timestamp := currentTime.Unix()
			dayTS := timestamp - (timestamp % 86400)
			//todo only one brdHdr is required for a cluster
			cids, brdHdrs, _ := ctc.GetDailyBrdMail(&bind.CallOpts{}, clusterId, uint64(dayTS))
			//fmt.Println(cids, clusterId)
			for k := 0; k < len(cids); k++ {
				cid := cids[k]
				brdHdr := brdHdrs[k]
				hdr := broadcast.Header{PointToG1(brdHdr.C0), PointToG2(brdHdr.C0p), PointToG1(brdHdr.C1)}
				//ptr := my.Domains[dmId].SK
				sk := created.Domains[dmId].SK
				//fmt.Println(sk.Di.String())
				beKp := sk.Decrypt(created.Domains[dmId].Clusters[clusterId], hdr, created.Domains[dmId].PKs)
				//V := created.Domains[dmId].PKs.V
				//fmt.Println(i, "beKp", created.Domains[dmId].Clusters[clusterId], dmId, beKp.String()[:30], V.String()[:30])
				os.MkdirAll("../users/"+created.Psid, os.ModePerm)
				GetIPFSClient().Get(cid, "../users/"+created.Psid+"/")
				file, _ := os.Open("../users/" + created.Psid + "/" + cid)
				content, _ := io.ReadAll(file)
				decRes, _ := aes.Decrypt(string(content), beKp.Marshal()[:32])
				if isPrintable(decRes) {
					fmt.Println(created.Psid + " Read email: \033[34m" + decRes + "\033[0m")
				} else {
					fmt.Println(created.Psid + " is \033[31m not \033[0m in the cluster.")
				}

			}

		}
	}

}

// construct a transaction
func TransactValue(client *ethclient.Client, privatekey string, toAddr common.Address, value *big.Int) *types.Receipt {
	key, _ := crypto.HexToECDSA(privatekey)
	publicKey := key.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)
	tx := types.NewTransaction(nonce, toAddr, value, uint64(900719925), big.NewInt(20000000000), nil)
	chainID, _ := client.ChainID(context.Background())
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	client.SendTransaction(context.Background(), signedTx)
	// 等待交易被矿工确认
	receipt, _ := bind.WaitMined(context.Background(), client, signedTx)
	return receipt
}

// deploy contract and obtain abi interface and bin of source code
func Deploy(client *ethclient.Client, contract_name string, auth *bind.TransactOpts) (common.Address, *types.Transaction) {
	abiBytes, _ := os.ReadFile("../compile/contract/" + contract_name + ".abi")
	bin, _ := os.ReadFile("../compile/contract/" + contract_name + ".bin")
	parsedABI, _ := abi.JSON(strings.NewReader(string(abiBytes)))
	address, tx, _, _ := bind.DeployContract(auth, parsedABI, common.FromHex(string(bin)), client)
	receipt, _ := bind.WaitMined(context.Background(), client, tx)
	fmt.Printf("Basics.sol deployed! Address: %s Gas used: %d\n", address.Hex(), receipt.GasUsed)
	return address, tx
}

// construct a transaction
func Transact(client *ethclient.Client, privatekey string, value *big.Int, ctc *contract.Contract, para []interface{}) interface{} {
	key, _ := crypto.HexToECDSA(privatekey)
	publicKey := key.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	chainID, err := client.ChainID(context.Background())
	auth, _ := bind.NewKeyedTransactorWithChainID(key, chainID)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = value
	auth.GasLimit = uint64(900719925)       //gasLimit
	auth.GasPrice = big.NewInt(20000000000) //gasPrice
	if ctc == nil {
		return auth
	}
	var f interface{}
	args := []interface{}{auth}
	for i := 1; i < len(para); i++ {
		args = append(args, para[i])
	}
	//fmt.Println(args)
	switch para[0] {
	//case "HashToG1":
	//	f = ctc.HashToG1
	case "Register":
		f = ctc.Register
	case "MailTo":
		f = ctc.MailTo
	case "BcstTo":
		f = ctc.BcstTo
	case "RegDomain":
		f = ctc.RegDomain
	case "RegCluster":
		f = ctc.RegCluster
	case "GetBrdEncPrivs":
		f = ctc.GetBrdEncPrivs
	case "LinkTmpPsid":
		f = ctc.LinkTmpPsid
	case "GetTmpPsid":
		f = ctc.GetTmpPsid
		//case "SplitAt":
		//	f = ctc.SplitAt
	}

	// 获取函数的反射值
	funcValue := reflect.ValueOf(f)
	// 构造参数列表
	var params []reflect.Value
	for _, arg := range args {
		params = append(params, reflect.ValueOf(arg))
	}
	//fmt.Println(params)

	// 调用函数
	resultValues := funcValue.Call(params)
	//fmt.Println(resultValues[0].Kind(), resultValues[0].Type())
	tx := resultValues[0].Interface().(*types.Transaction)
	receipt, _ := bind.WaitMined(context.Background(), client, tx)
	//fmt.Printf("HashToG1() Gas used: %d\n", receipt.GasUsed)
	fmt.Printf("%v Gas used: %d\n", para[0], receipt.GasUsed)
	return receipt

	return auth
}
