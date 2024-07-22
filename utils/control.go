package utils

import "C"
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

	ioutil.WriteFile(GetGoModPath()+"/compile/contract/addr.txt", []byte(address.String()), 0644)
	//fmt.Println(GetGoModPath() + "/compile/contract/addr.txt")
	fmt.Println("=============================register psids and public keys=====================")
	//Users register their public keys (A B)
	//names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy"}
	names := []string{"Bob", "Alice", "Charlie", "Emily", "Alexander", "Sophia", "Benjamin", "Olivia", "James", "Peggy", "Isabella", "Jacob", "Ava", "Matthew", "Mia", "Daniel", "Abigail", "Ethan", "Harper", "Max", "Amelia", "Ryan", "Evelyn", "Nathan", "Elizabeth", "Samuel", "Charlotte", "Christopher", "Grace", "Jonathan", "Lily", "Gabriel", "Ella", "Andrew", "Avery", "Joshua", "Sofia", "Anthony", "Scarlett", "Caleb", "Victoria", "Logan", "Madison", "Isaac", "Eleanor", "Lucas", "Hannah", "Owen", "Addison", "Dylan", "Zoe", "Jack", "Penelope", "Luke", "Layla", "Jeremiah", "Natalie", "Isaiah", "Audrey", "Carter", "Leah", "Josiah", "Savannah", "Julian", "Brooklyn", "Wyatt", "Stella", "Hunter", "Claire", "Levi", "Skylar", "Christian", "Maya", "Eli", "Paisley", "Lincoln", "Anna", "Jordan", "Caroline", "Charles", "Eliana", "Thomas", "Ruby", "Aaron", "Aria", "Connor", "Aurora", "Cameron", "Naomi", "Adrian", "Valentina", "Landon", "Alexa", "Gavin", "Lydia", "Evan", "Piper", "Sebastian", "Ariana", "Cooper", "Sadie"}
	names = names[:10]
	users := make([]User, len(names))
	for i := 0; i < len(names); i++ {
		a, _ := rand.Int(rand.Reader, bn256.Order)
		b, _ := rand.Int(rand.Reader, bn256.Order)
		A := new(bn256.G1).ScalarBaseMult(a)
		B := new(bn256.G1).ScalarBaseMult(b)
		privatekey := GetENV("PRIVATE_KEY_" + strconv.Itoa(i%10+1))
		key, _ := crypto.HexToECDSA(ReverseString(privatekey))
		publicKey := key.Public()
		publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
		addr := crypto.PubkeyToAddress(*publicKeyECDSA)
		para := []interface{}{"Register", names[i], contract.EmailPK{G1ToPoint(A), G1ToPoint(B), ether01, addr, make([]contract.EmailG1Point, 0)}, ""}
		_ = Transact(client, privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
		var domains = make(map[string]Domain)
		users[i] = User{names[i], a, b, A, B, privatekey, addr.String(), domains}
		//users[i] = ResolveUser(ctc, "Emily", a, b, "", "", i)
	}
	//users generate their BIP32 child keys
	InitBIP32Wallet(client, users)
	return users, client, ctc
}
func Oto(client *ethclient.Client, ctc *contract.Contract, sender User, msg []byte, to User, recs []string) string {
	m, _ := rand.Int(rand.Reader, bn256.Order)
	key := new(bn256.G1).ScalarBaseMult(m)

	pkRes, _ := ctc.GetPK(&bind.CallOpts{}, to.Psid)
	sa := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)})
	r, _ := rand.Int(rand.Reader, bn256.Order)
	c1 := new(bn256.G1).ScalarBaseMult(r) // c1 = r * G
	c2 := new(bn256.G1).Add(new(bn256.G1).ScalarMult(sa.S, r), key)
	ct, _ := aes.Encrypt(msg, key.Marshal()[:32])
	cid := IPFSUpload(ct)
	mail := contract.EmailMail{contract.EmailStealthPub{G1ToPoint(sa.R), G1ToPoint(sa.S)}, G1ToPoint(c1), G1ToPoint(c2)}
	para := []interface{}{"Oto", mail, cid, recs}
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
		dir := GetGoModPath() + "/users/" + my.Psid + "/"
		EnsureDir(dir)
		GetIPFSClient().Get(cid2Flag[0], dir)
		file, _ := os.Open(dir + cid2Flag[0])
		content, _ := io.ReadAll(file)
		decRes := string(content)
		//  ElGamal decrypt
		c1pNeg := new(bn256.G1).Neg(PointToG1(dayMails[i].C1))
		c2p := PointToG1(dayMails[i].C2)
		keyp := new(bn256.G1).Add(c2p, new(bn256.G1).ScalarMult(c1pNeg, sp))
		decRes, _ = aes.Decrypt(decRes, keyp.Marshal()[:32])
		//}
		fmt.Println("Email content (read): \033[34m" + decRes + "\033[0m")
	}
}
func RegCluster(client *ethclient.Client, ctc *contract.Contract, from User, clsId string, S []uint32) {
	para := []interface{}{"RegCluster", clsId, S}
	//fmt.Printf("%v", para)
	_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
}

func RegDomain(client *ethclient.Client, ctc *contract.Contract, from User, cpk broadcast.PKs, brdPrivs []broadcast.SK, users []User) {

	encPrivs := make([]contract.EmailStealthEncPriv, len(users))
	psids := make([]string, len(brdPrivs))
	for i := 0; i < len(users); i++ {
		pkRes, _ := ctc.GetPK(&bind.CallOpts{}, users[i].Psid)
		sa := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)})
		// ElGamal encrypted using a stealth address
		r, _ := rand.Int(rand.Reader, bn256.Order)
		encPrivs[i].C1 = G1ToPoint(new(bn256.G1).ScalarBaseMult(r))
		encPrivs[i].C2 = G1ToPoint(new(bn256.G1).Add(new(bn256.G1).ScalarMult(sa.S, r), &brdPrivs[i+1].Di))
		encPrivs[i].R = G1ToPoint(sa.R)
		encPrivs[i].S = G1ToPoint(sa.S)
		psids[i] = users[i].Psid
		//fmt.Println("------------", psids[i], cts[i].C1, len(cts), len(psids))
	}
	//fmt.Println(len(cpk.PArr), len(cpk.QArr), len(users)) //2n+1,n+1,n
	para := []interface{}{"RegDomain", cpk.DmId, G1ArrToPoints(cpk.PArr), G2ArrToPoints(cpk.QArr),
		G1ToPoint(&cpk.V), encPrivs, psids}
	//fmt.Printf("%v", para)
	_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
}

//	func DownloadPKs(ctc *contract.Contract, clsId string) broadcast.SK {
//		dmId := strings.Split(clsId, "@")[1]
//		pArr, qArr, v, _ := ctc.GetBrdPKs(&bind.CallOpts{}, dmId)
//		S, _ := ctc.GetS(&bind.CallOpts{}, clsId)
//		//fmt.Println(S)
//		brdPks := broadcast.PKs{PointsToG1(pArr), PointsToG2(qArr), *PointToG1(v), dmId}
//	}

func ResolveTmpUser(ctc *contract.Contract, psid string, Aa *big.Int, Bb *big.Int, wallet string, idx int) []User {
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

		privatekey := GetENV("PRIVATE_KEY_" + strconv.Itoa(idx%10+1))
		user := User{tmpPsids[i], a, b, new(bn256.G1).ScalarBaseMult(a),
			new(bn256.G1).ScalarBaseMult(b), privatekey, wallet, domains}
		for k := 0; k < len(DomainIds); k++ {
			dmId := DomainIds[k].DmId
			index := DomainIds[k].Index
			encPrivs, _ := ctc.GetBrdEncPrivs(&bind.CallOpts{}, dmId, user.Psid)

			priv := stealth.ResolvePriv(stealth.SecretKey{user.Aa, user.Bb}, stealth.StealthPub{PointToG1(encPrivs.R), PointToG1(encPrivs.S)})
			//ElGamal Decryption
			c1pNeg := new(bn256.G1).Neg(PointToG1(encPrivs.C1))
			myBrdPriv := new(bn256.G1).Add(PointToG1(encPrivs.C2), new(bn256.G1).ScalarMult(c1pNeg, priv))
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
			user.Domains[dmId] = Domain{brdPks, broadcast.SK{int(index.Int64()), *myBrdPriv}, clsIds}
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

		encPrivs, _ := ctc.GetBrdEncPrivs(&bind.CallOpts{}, dmId, user.Psid)

		priv := stealth.ResolvePriv(stealth.SecretKey{user.Aa, user.Bb}, stealth.StealthPub{PointToG1(encPrivs.R), PointToG1(encPrivs.S)})
		//ElGamal Decryption
		c1pNeg := new(bn256.G1).Neg(PointToG1(encPrivs.C1))
		myBrdPriv := new(bn256.G1).Add(PointToG1(encPrivs.C2), new(bn256.G1).ScalarMult(c1pNeg, priv))

		var clsIds = map[string][]uint32{}

		clsIdsDL, _ := ctc.GetMyClusters(&bind.CallOpts{}, dmId)
		for j := 0; j < len(clsIdsDL); j++ {
			clsId := clsIdsDL[j]
			S, _ := ctc.GetS(&bind.CallOpts{}, clsId)
			clsIds[clsId] = S
		}
		user.Domains[dmId] = Domain{brdPks, broadcast.SK{int(index.Int64()), *myBrdPriv}, clsIds}
	}
	//fmt.Printf("user %v\n", user)
	return user
}

func BroadcastTo(client *ethclient.Client, ctc *contract.Contract, sender User, msg []byte, clusterId string) string {
	// todo download brdPKs
	dmId := strings.Split(clusterId, "@")[1]
	brdPKs := sender.Domains[dmId].PKs
	hdr, beK := brdPKs.Encrypt(sender.Domains[dmId].Clusters[clusterId])
	//fmt.Println("beK", clusterId, sender.Domains[dmId].Clusters[clusterId], beK.String()[:30], "hdr", brdPKs.V.String()[:30])
	//fmt.Println("beK.String()", beK.String()[:30], "hdr:", hdr.C0.String())
	ct, _ := aes.Encrypt(msg, beK.Marshal()[:32])
	fmt.Println("encrypted email to broadcast:", ct)
	// Bob uploads encrypted email content to IPFS
	cid, _ := GetIPFSClient().Add(strings.NewReader(ct))
	fmt.Println("broadcast mail IPFS link:", cid)
	//x, _ := rand.Int(rand.Reader, bn256.Order)
	r, _ := rand.Int(rand.Reader, bn256.Order)
	//senderIndex := sender.Domains[dmId].SK.I
	encPrivs, _ := ctc.GetBrdEncPrivs(&bind.CallOpts{}, dmId, sender.Psid)
	priv := stealth.ResolvePriv(stealth.SecretKey{sender.Aa, sender.Bb}, stealth.StealthPub{PointToG1(encPrivs.R), PointToG1(encPrivs.S)})
	SAPK := PointToG1(encPrivs.S)
	//fmt.Printf("encPrivs.R: %v\n", PointToG1(encPrivs.R).String())
	Cp := new(bn256.G1).ScalarMult(&brdPKs.PArr[0], r)
	hash := sha256.Sum256([]byte(SAPK.String() + Cp.String()))
	c := new(big.Int).SetBytes(hash[:])

	//fmt.Printf("%v", stealth.SecretKey{sender.Aa, sender.Bb})
	ctilde := new(big.Int).Sub(r, new(big.Int).Mul(priv, c))
	ctilde = new(big.Int).Mod(ctilde, bn256.Order)
	//fmt.Printf("%v, %d, %v\n", senderIndex, priv, SAPK.String())
	proof := contract.EmailPi{G1ToPoint(Cp), c, ctilde}
	clusterRecivers := contract.EmailBcstHeader{G1ToPoint(hdr.C0), G1ToPoint(hdr.C1), G2ToPoint(hdr.C0p)}
	// schnorr sigma protocol verification
	// fmt.Printf("point %v %v %v %v\n", proof, cid, clusterId, clusterRecivers)
	para := []interface{}{"BcstTo", clusterRecivers, clusterId, proof, cid, sender.Psid}
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
			for k := 0; k < len(cids); k++ {
				cid := cids[k]
				brdHdr := brdHdrs[k]
				hdr := broadcast.Header{PointToG1(brdHdr.C0), PointToG2(brdHdr.C0p), PointToG1(brdHdr.C1)}
				//ptr := my.Domains[dmId].SK
				sk := created.Domains[dmId].SK
				beKp := sk.Decrypt(created.Domains[dmId].Clusters[clusterId], hdr, created.Domains[dmId].PKs)
				//V := created.Domains[dmId].PKs.V
				//fmt.Println(i, "beKp", created.Domains[dmId].Clusters[clusterId], dmId, beKp.String()[:30])
				dir := GetGoModPath() + "/users/" + created.Psid + "/"
				EnsureDir(dir)
				//os.MkdirAll(dir, os.ModePerm)
				GetIPFSClient().Get(cid, dir)
				file, _ := os.Open(dir + cid)
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
func Reward(client *ethclient.Client, ctc *contract.Contract, sender User, dmId string) {
	para := []interface{}{"Reward", dmId}
	_ = Transact(client, sender.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
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
	abiBytes, _ := os.ReadFile(GetGoModPath() + "/compile/contract/" + contract_name + ".abi")
	bin, _ := os.ReadFile(GetGoModPath() + "/compile/contract/" + contract_name + ".bin")
	parsedABI, _ := abi.JSON(strings.NewReader(string(abiBytes)))
	address, tx, _, _ := bind.DeployContract(auth, parsedABI, common.FromHex(string(bin)), client)
	receipt, _ := bind.WaitMined(context.Background(), client, tx)
	fmt.Printf("\n\nContract is deployed! Address: %s Gas used: %d\n", address.Hex(), receipt.GasUsed)
	return address, tx
}

// construct a transaction
func Transact(client *ethclient.Client, privatekey string, value *big.Int, ctc *contract.Contract, para []interface{}) interface{} {
	key, _ := crypto.HexToECDSA(privatekey)
	publicKey := key.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)

	chainID, _ := client.ChainID(context.Background())
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
	case "Oto":
		f = ctc.Oto
	case "BcstTo":
		f = ctc.BcstTo
	case "RegDomain":
		f = ctc.RegDomain
	case "RegCluster":
		f = ctc.RegCluster
	case "GetBrdEncPrivs":
		f = ctc.GetBrdEncPrivs
	// case "LinkTmpPsid":
	// 	f = ctc.LinkTmpPsid
	case "GetTmpPsid":
		f = ctc.GetTmpPsid
	case "Reward":
		f = ctc.Reward
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
	// fmt.Println(para[0].(string) == "Register", strings.Contains(para[1].(string), "Alice"))
	if para[0].(string) != "Register" || (para[0].(string) == "Register" && strings.Contains(para[1].(string), "Alice")) {
		fmt.Printf("%v Gas used: %d\n", para[0], receipt.GasUsed)
	}
	// fmt.Printf("%v Gas used: %d\n", para[0], receipt.GasUsed)
	return receipt
}
