// go与区块链交互需要的函数
package utils

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"email/compile/contract"
	"email/crypto/aes"
	"email/crypto/broadcast"
	"email/crypto/stealth"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fentec-project/bn256"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joho/godotenv"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// deploy contract and obtain abi interface and bin of source code
func Deploy(client *ethclient.Client, contract_name string, auth *bind.TransactOpts) (common.Address, *types.Transaction) {

	abiBytes, err := os.ReadFile("compile/contract/" + contract_name + ".abi")
	if err != nil {
		log.Fatalf("Failed to read ABI file: %v", err)
	}

	bin, err := os.ReadFile("compile/contract/" + contract_name + ".bin")
	if err != nil {
		log.Fatalf("Failed to read BIN file: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(string(bin)), client)
	if err != nil {
		log.Fatalf("Failed to deploy contract: %v", err)
	}

	receipt, _ := bind.WaitMined(context.Background(), client, tx)
	fmt.Printf("Basics.sol deployed! Address: %s Gas used: %d\n", address.Hex(), receipt.GasUsed)
	//fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
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
	case "RegGroup":
		f = ctc.RegGroup
	case "RegDomain":
		f = ctc.RegDomain
	case "RetrBrdPrivs":
		f = ctc.RetrBrdPrivs
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
	//// 处理返回值
	//var result int
	//if len(resultValues) > 0 {
	//	// 检查返回值的有效性
	//	if resultValues[0].Kind() == reflect.Int {
	//		result = int(resultValues[0].Int())
	//		fmt.Println("Result:", result)
	//	} else {
	//		fmt.Println("Function did not return an int")
	//	}
	//}

	//tx3, _ := fn(auth, "11223")
	//fmt.Printf("Onchain %v result: %v\n", fn, tx3)

	return auth
}

// construct a transaction
func TransactValue(client *ethclient.Client, privatekey string, toAddr common.Address, value *big.Int) *types.Receipt {
	key, _ := crypto.HexToECDSA(privatekey)
	publicKey := key.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	tx := types.NewTransaction(nonce, toAddr, value, uint64(900719925), big.NewInt(20000000000), nil)
	// 获取网络ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}
	// 等待交易被矿工确认
	receipt, err := bind.WaitMined(context.Background(), client, signedTx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction mining: %v", err)
	}
	return receipt
}

func GetENV(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	return os.Getenv(key)
}

func Hash2G1(msg string) *bn256.G1 {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(msg))
	v := hash.Sum(nil)
	return new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(v))

}
func GetAddr(privatekey string) common.Address {
	senderkey, _ := crypto.HexToECDSA(privatekey)
	senderPKECDSA, _ := senderkey.Public().(*ecdsa.PublicKey)
	senderAddr := crypto.PubkeyToAddress(*senderPKECDSA)
	return senderAddr
}

func InitBIP32Wallet(client *ethclient.Client, users []User) {
	for j := 0; j < len(users); j++ {
		envMap, _ := godotenv.Read(".env")
		ether := big.NewInt(1000000000000000000)
		prvKeySeed := GetENV("MASTER_KEY_" + users[j].Psid)
		var msk *bip32.Key
		if prvKeySeed == "" {
			seed, _ := bip32.NewSeed()
			msk, _ = bip32.NewMasterKey(seed)
		} else {
			bts, _ := hex.DecodeString(prvKeySeed)
			msk, _ = bip32.Deserialize(bts)
		}
		mskStr := hex.EncodeToString(msk.Key)
		masterBalance, _ := client.BalanceAt(context.Background(), GetAddr(mskStr), nil)
		if masterBalance.Cmp(big.NewInt(1).Mul(ether, big.NewInt(100))) > 0 {
			fmt.Printf("masterBalance: %s ether\n", masterBalance.Div(masterBalance, ether))
			return
		}
		s, _ := msk.Serialize()
		envMap["MASTER_KEY_"+users[j].Psid] = hex.EncodeToString(s)

		recipt := TransactValue(client, users[j].Privatekey, GetAddr(mskStr), big.NewInt(1).Mul(big.NewInt(1000), ether)) //1000ETH
		if j == 0 {
			fmt.Println("TransactValue recipt gas used:", recipt.GasUsed)
		}

		chdKeys := map[int]*bip32.Key{}
		for i := 1; i < 4; i++ {
			chdKeys[i], _ = msk.NewChildKey(uint32(i))
			childKeyStr := hex.EncodeToString(chdKeys[i].Key)
			TransactValue(client, users[j].Privatekey, GetAddr(childKeyStr), big.NewInt(1).Mul(big.NewInt(1000), ether)) //1000ETH
			envMap[users[j].Psid+"_"+strconv.Itoa(i)] = childKeyStr
		}
		godotenv.Write(envMap, ".env")
	}

}
func G1ArrToPoints(points []bn256.G1) []contract.EmailG1Point {
	arr := make([]contract.EmailG1Point, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = G1ToPoint(&points[i])
	}
	return arr
}
func G2ArrToPoints(points []bn256.G2) []contract.EmailG2Point {
	arr := make([]contract.EmailG2Point, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = G2ToPoint(&points[i])
	}
	return arr
}
func G1ToPoint(point *bn256.G1) contract.EmailG1Point {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	pointBytes := point.Marshal()
	x := new(big.Int).SetBytes(pointBytes[:32])
	y := new(big.Int).SetBytes(pointBytes[32:64])

	g1Point := contract.EmailG1Point{
		X: x,
		Y: y,
	}
	return g1Point
}

func G2ToPoint(point *bn256.G2) contract.EmailG2Point {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	pointBytes := point.Marshal()
	//fmt.Println(point.Marshal())

	// Create big.Int for X and Y coordinates
	a1 := new(big.Int).SetBytes(pointBytes[:32])
	a2 := new(big.Int).SetBytes(pointBytes[32:64])
	b1 := new(big.Int).SetBytes(pointBytes[64:96])
	b2 := new(big.Int).SetBytes(pointBytes[96:128])

	g2Point := contract.EmailG2Point{
		X: [2]*big.Int{a1, a2},
		Y: [2]*big.Int{b1, b2},
	}
	return g2Point
}
func PointsToG1(points []contract.EmailG1Point) []bn256.G1 {
	arr := make([]bn256.G1, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = *PointToG1(points[i])
	}
	return arr
}
func PointsToG2(points []contract.EmailG2Point) []bn256.G2 {
	arr := make([]bn256.G2, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = *PointToG2(points[i])
	}
	return arr
}
func PointToG1(point contract.EmailG1Point) *bn256.G1 {
	combinedByteArray := make([]byte, 64)
	point.X.FillBytes(combinedByteArray[:32])
	point.Y.FillBytes(combinedByteArray[32:])

	g1 := new(bn256.G1)
	g1.Unmarshal(combinedByteArray)
	return g1

}
func PointToG2(point contract.EmailG2Point) *bn256.G2 {
	combinedByteArray := make([]byte, 128)
	point.X[0].FillBytes(combinedByteArray[:32])
	point.X[1].FillBytes(combinedByteArray[32:64])
	point.Y[0].FillBytes(combinedByteArray[64:96])
	point.Y[1].FillBytes(combinedByteArray[96:128])
	g2 := new(bn256.G2)
	g2.Unmarshal(combinedByteArray)
	return g2

}

var ipfs *shell.Shell

func GetIPFSClient() *shell.Shell {
	if ipfs == nil {
		ipfs = shell.NewShell("localhost:5001")
	}
	return ipfs
}
func IPFSUpload(msg string) string {
	sh := GetIPFSClient()
	cid, _ := sh.Add(strings.NewReader(msg))
	fmt.Println("Mail IPFS link:", cid)
	return cid
}
func MailTo(client *ethclient.Client, ctc *contract.Contract, sender User, key *bn256.G1, msg []byte, to User, recs []string) string {
	pkRes, _ := ctc.DownloadPK(&bind.CallOpts{}, to.Psid)
	sa := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)})
	r, _ := rand.Int(rand.Reader, bn256.Order)
	c1 := new(bn256.G1).ScalarBaseMult(r) // c1 = r * G
	c2 := new(bn256.G1).Add(new(bn256.G1).ScalarMult(sa.S, r), key)
	ct, _ := aes.Encrypt(msg, key.Marshal()[:32])
	cid := IPFSUpload(ct) + "||0"
	para := []interface{}{"MailTo", contract.EmailStealthAddrPub{G1ToPoint(sa.R), G1ToPoint(sa.S)}, contract.EmailElGamalCT{G1ToPoint(c1), G1ToPoint(c2)}, cid, append(recs, to.Psid)}
	_ = Transact(client, sender.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)

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
			stealth.StealthAddrPub{PointToG1(dayMails[i].Pub.R), PointToG1(dayMails[i].Pub.S)})
		cid2Flag := strings.Split(cids[i], "||")
		fmt.Println(cid2Flag)
		GetIPFSClient().Get(cid2Flag[0], "./"+my.Psid+"/")
		file, _ := os.Open("./" + my.Psid + "/" + cid2Flag[0])
		content, _ := io.ReadAll(file)
		decRes := string(content)
		if cid2Flag[1] == "0" {
			c1pNeg := new(bn256.G1).Neg(PointToG1(dayMails[i].Ct.C1))
			c2p := PointToG1(dayMails[i].Ct.C2)
			keyp := new(bn256.G1).Add(c2p, new(bn256.G1).ScalarMult(c1pNeg, sp))
			decRes, _ = aes.Decrypt(decRes, keyp.Marshal()[:32])
		}
		fmt.Println("Email content (read):", decRes)
	}

}
func RegDomain(client *ethclient.Client, ctc *contract.Contract, from User, name string, S []uint32) {
	para := []interface{}{"RegDomain", name, S}
	_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
}

func RegGroup(client *ethclient.Client, ctc *contract.Contract, from User, cpk broadcast.PKs, privs []broadcast.SK, to []User) {
	c1 := make([]bn256.G1, len(privs))
	c2 := make([]bn256.G1, len(privs))
	names := make([]string, len(privs))
	for i := 0; i < len(to); i++ {
		pkRes, _ := ctc.DownloadPK(&bind.CallOpts{}, to[i].Psid)
		//sa := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)})
		r, _ := rand.Int(rand.Reader, bn256.Order)
		c1[i] = *new(bn256.G1).ScalarBaseMult(r)
		// todo whether a stealth address is needed?
		c2[i] = *new(bn256.G1).Add(new(bn256.G1).ScalarMult(PointToG1(pkRes.A), r), &privs[i+1].Di)
		//fmt.Println(c1[i].String())
		names[i] = to[i].Psid
	}

	//fmt.Println(len(cpk.PArr), len(cpk.QArr), len(to)) //2n+1,n+1,n
	para := []interface{}{"RegGroup", cpk.GrpId, G1ArrToPoints(cpk.PArr), G2ArrToPoints(cpk.QArr),
		G1ToPoint(&cpk.V), G1ArrToPoints(c1), G1ArrToPoints(c2), names}
	_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
}
func DownloadAndResolvePriv(ctc *contract.Contract, my User, groupname string) broadcast.SK {
	index, c1, c2, _ := ctc.RetrBrdPrivs(&bind.CallOpts{}, groupname, my.Psid)
	c1pNeg := new(bn256.G1).Neg(PointToG1(c1))
	myBrdPriv := new(bn256.G1).Add(PointToG1(c2), new(bn256.G1).ScalarMult(c1pNeg, my.Aa))
	//
	//fmt.Printf("Secret key of broadcast email: %v %v\n", myBrdPriv.String(), index)
	return broadcast.SK{
		int(index.Int64()) + 1, *myBrdPriv,
	}
}

func BroadcastTo(client *ethclient.Client, ctc *contract.Contract, sender User, msg []byte, domainId string) string {
	// todo download cpk
	grpId := strings.Split(domainId, "@")[1]
	brdPKs := sender.Groups[grpId].PKs
	hdr, beK := brdPKs.Encrypt(sender.Groups[grpId].Domains[domainId])
	//fmt.Println("beK.String()", beK.String()[:30])
	ct, _ := aes.Encrypt(msg, beK.Marshal()[:32])
	fmt.Println("encrypted email to broadcast:", ct)
	// Bob uploads encrypted email content to IPFS
	cid, _ := GetIPFSClient().Add(strings.NewReader(ct))
	fmt.Println("broadcast mail IPFS link:", cid)
	x, _ := rand.Int(rand.Reader, bn256.Order)
	//x = big.NewInt(1)
	senderIndex := sender.Groups[grpId].SK.I
	domainRecivers := contract.EmailBrdcastHeader{G1ToPoint(hdr.C0), G1ToPoint(hdr.C1), G2ToPoint(hdr.C0p)}
	ptr := sender.Groups[grpId].SK.Di
	proof := contract.EmailDomainProof{G1ToPoint(new(bn256.G1).ScalarMult(&ptr, x)),
		G2ToPoint(&brdPKs.QArr[senderIndex]), G1ToPoint(new(bn256.G1).ScalarMult(&brdPKs.V, x))}
	//e(skipws,g2)= e(pki,vpows)
	para := []interface{}{"BcstTo", domainRecivers, domainId, proof, cid}
	_ = Transact(client, sender.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)
	return cid
}

func ReadBrdMail(ctc *contract.Contract, my User, domainId string) {
	//contract.EmailBrdcastCT{G1ToPoint(hdr.C0), G1ToPoint(hdr.C1)}
	grpId := strings.Split(domainId, "@")[1]
	currentTime := time.Now()
	timestamp := currentTime.Unix()
	dayTS := timestamp - (timestamp % 86400)
	//todo only one brdHdr is required for a domain
	cids, brdHdrs, _ := ctc.GetDailyBrdMail(&bind.CallOpts{}, domainId, uint64(dayTS))
	for i := 0; i < len(cids); i++ {
		cid := cids[i]
		brdHdr := brdHdrs[i]
		hdr := broadcast.Header{
			PointToG1(brdHdr.C0),
			PointToG2(brdHdr.C0p),
			PointToG1(brdHdr.C1),
		}
		ptr := my.Groups[grpId].SK
		//fmt.Println(i, "my.Groups[grpId].SK", ptr.Di, ptr.Di.String()[:30])
		beKp := ptr.Decrypt(my.Groups[grpId].Domains[domainId], hdr, my.Groups[grpId].PKs)
		GetIPFSClient().Get(cid, "./"+my.Psid+"/")
		file, _ := os.Open("./" + my.Psid + "/" + cid)
		content, _ := io.ReadAll(file)

		decRes, _ := aes.Decrypt(string(content), beKp.Marshal()[:32])
		fmt.Println("Broadcast Email content (read):", decRes)
	}

}
