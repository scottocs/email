// go与区块链交互需要的函数
package utils

import (
	"context"
	"crypto/ecdsa"
	"email/compile/contract"
	"encoding/hex"
	"fmt"
	"github.com/fentec-project/bn256"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
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
	fmt.Printf("Basics.sol deployed! Address: %s\n", address.Hex())
	fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
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
	case "HashToG1":
		f = ctc.HashToG1
	case "UploadPK":
		f = ctc.UploadPK
	case "MailTo":
		f = ctc.MailTo
	case "BrdcastTo":
		f = ctc.BrdcastTo
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

func InitBIP32Wallet(client *ethclient.Client, privatekey1 string) {

	envMap, _ := godotenv.Read(".env")

	ether := big.NewInt(1000000000000000000)

	bts, _ := hex.DecodeString(GetENV("MASTER_KEY"))
	msk, _ := bip32.Deserialize(bts)
	mskStr := ""
	if msk != nil {
		mskStr = hex.EncodeToString(msk.Key)
		masterBalance, _ := client.BalanceAt(context.Background(), GetAddr(mskStr), nil)
		if masterBalance.Cmp(big.NewInt(1).Mul(ether, big.NewInt(100))) > 0 {
			fmt.Printf("masterBalance: %s ether\n", masterBalance.Div(masterBalance, ether))
			return
		}
	} else {
		seed, _ := bip32.NewSeed()
		msk, _ = bip32.NewMasterKey(seed)
		mskStr = hex.EncodeToString(msk.Key)
		s, _ := msk.Serialize()
		envMap["MASTER_KEY"] = hex.EncodeToString(s)
	}
	fmt.Println("mskStr", mskStr)
	//envMap["PRIVATE_KEY_100"] = mskStr

	recipt := TransactValue(client, privatekey1, GetAddr(mskStr), big.NewInt(1).Mul(big.NewInt(1000), ether)) //1000ETH
	fmt.Println("TransactValue recipt gas used:", recipt.GasUsed)
	chdKeys := map[int]*bip32.Key{}
	for i := 1; i <= 10; i++ {
		chdKeys[i], _ = msk.NewChildKey(uint32(i))
		childKeyStr := hex.EncodeToString(chdKeys[i].Key)
		TransactValue(client, privatekey1, GetAddr(childKeyStr), big.NewInt(1).Mul(big.NewInt(1000), ether)) //1000ETH
		envMap["PRIVATE_KEY_"+strconv.Itoa(i+10)] = childKeyStr
		//bal, _ := client.BalanceAt(context.Background(), GetAddr(childKeyStr), nil)
		//fmt.Printf("%dthmasterBalance: %s ether\n", i+10, bal.Div(bal, ether))
	}
	//senderBalance, _ := client.BalanceAt(context.Background(), GetAddr(privatekey1), nil)
	//fmt.Printf("senderBalance: %s ether\n", senderBalance.Div(senderBalance, ether))
	godotenv.Write(envMap, ".env")

	//"e5c2c2a3f5b7aa82c790d131b209443befd377c84d490f2cf4d5ed408f008ae0"
	//addressBytes, _ := hex.DecodeString("41A5BC8ecbFad34e14af34C5E690A5355C861fA6")
	//common.BytesToAddress(addressBytes)
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

func PointToG1(point contract.EmailG1Point) *bn256.G1 {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	//new(bn256.G1).Un
	//func (e *GT) Unmarshal(m []byte) ([]byte, error) {
	//pointBytes := point.Marshal()
	combinedByteArray := make([]byte, 64)
	point.X.FillBytes(combinedByteArray[:32])
	point.Y.FillBytes(combinedByteArray[32:])

	g1 := new(bn256.G1)
	g1.Unmarshal(combinedByteArray)
	return g1

}
