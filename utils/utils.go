// go与区块链交互需要的函数
package utils

import (
	"context"
	"crypto/ecdsa"
	"email/compile/contract"
	"email/crypto/broadcast"
	"email/crypto/stealth"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fentec-project/bn256"
	"github.com/joho/godotenv"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/rand"
	"log"
	"math/big"
	rand2 "math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"unicode"
)

func InitBIP32Wallet(client *ethclient.Client, users []User) {
	for j := 0; j < len(users); j++ {
		//envMap, _ := godotenv.Read(findEnvFile()".env")
		envMap := GetAllEnv()
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
		//godotenv.Write(envMap, ".env")
		WriteAllEnv(envMap)
	}

}

func GetAddr(privatekey string) common.Address {
	senderkey, _ := crypto.HexToECDSA(privatekey)
	senderPKECDSA, _ := senderkey.Public().(*ecdsa.PublicKey)
	senderAddr := crypto.PubkeyToAddress(*senderPKECDSA)
	return senderAddr
}

func GetAddrFromPK(privatekey string) common.Address {
	senderkey, _ := crypto.HexToECDSA(privatekey)
	senderPKECDSA, _ := senderkey.Public().(*ecdsa.PublicKey)
	senderAddr := crypto.PubkeyToAddress(*senderPKECDSA)
	return senderAddr
}

func ReverseString(input string) string {
	runes := []rune(input)
	length := len(runes)

	for i := 0; i < length/2; i++ {
		runes[i], runes[length-i-1] = runes[length-i-1], runes[i]
	}
	return string(runes)
}

func GetENV(key string) string {
	//dir, _ := os.Getwd()
	err := godotenv.Load(GetGoModPath() + "/.env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	return os.Getenv(key)
}

func GetAllEnv() map[string]string {
	//dir, _ := os.Getwd()
	envMap, _ := godotenv.Read(GetGoModPath() + "/.env")
	return envMap
}
func WriteAllEnv(envMap map[string]string) {
	//dir, _ := os.Getwd()
	godotenv.Write(envMap, GetGoModPath()+"/.env")

}

func GetGoModPath() string {
	currentDir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return ""
		}
		currentDir = parentDir
	}
	return ""
}

func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func Hash2G1(msg string) *bn256.G1 {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(msg))
	v := hash.Sum(nil)
	return new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(v))
}

//	func Hash(msg string) string {
//		hash := sha3.NewLegacyKeccak256()
//		hash.Write([]byte(msg))
//		v := hash.Sum(nil)
//		return hex.Encode(v)
//	}
func shuffle(arr []uint32) {
	rand.Seed(uint64(time.Now().UnixNano()))
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func CreateTempCluster(client *ethclient.Client, ctc *contract.Contract, from User, psids []string) ([]User, string, []uint32) {
	users := make([]User, len(psids))

	rand2.Seed(time.Now().Unix())
	dmId := strconv.Itoa(rand2.Int())
	brdPks, brdPrivs := broadcast.Setup(len(psids), dmId)
	clusterId := "tmp@" + dmId
	//fmt.Println(rand2.Int(), clusterId, dmId)
	ClS := make([]uint32, len(psids))
	for i := 0; i < len(ClS); i++ {
		ClS[i] = uint32(i) + 1
	}
	//construct cluster from the provided psids
	shuffle(ClS)
	ClS = ClS[0 : len(ClS)/2]

	for i := 0; i < len(psids); i++ {
		//ClS[i] = uint32(i) + 1
		pkRes, _ := ctc.GetPK(&bind.CallOpts{}, psids[i])
		sa1 := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)}) //A
		sa2 := stealth.CalculatePub(stealth.PublicKey{PointToG1(pkRes.A), PointToG1(pkRes.B)}) //B

		var domain = make(map[string]Domain)
		//map[string][]uint32{clusterId: SEmily}
		var Smap = make(map[string][]uint32)
		Smap[clusterId] = ClS
		domain[dmId] = Domain{brdPks, brdPrivs[i+1], Smap}

		tmpPsid := psids[i] + "@" + clusterId
		privateKey := from.Privatekey //TODO Alice is user creator?

		users[i] = User{tmpPsid, nil, nil, sa1.S, sa2.S, privateKey, "", domain}
		//fmt.Println("11111111", clusterId, dmId, users[i].Domains[dmId].Clusters[clusterId])
		//ptr := users[i].Domains[dmId].SK.Di
		addr := common.BytesToAddress(([]byte)(users[i].Addr))
		para := []interface{}{"Register", tmpPsid, contract.EmailPK{G1ToPoint(sa1.S), G1ToPoint(sa2.S),
			big.NewInt(0), addr, []contract.EmailG1Point{G1ToPoint(sa1.R), G1ToPoint(sa2.R)}}, psids[i]}
		_ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para).(*types.Receipt)

		// para2 := []interface{}{"LinkTmpPsid", psids[i], tmpPsid}
		// _ = Transact(client, from.Privatekey, big.NewInt(0), ctc, para2).(*types.Receipt)
	}

	//ptr := users[0].Domains[dmId].SK.Di
	RegDomain(client, ctc, from, brdPks, brdPrivs, users)
	RegCluster(client, ctc, from, clusterId, ClS)
	//i := 0
	//for i = 0; i < len(ClS); i++ {
	//	if psids[ClS[i]-1] == "Alice" {
	//		fmt.Println("Alice is in the cluster: " + clusterId + " She is a receiver.")
	//		return users, clusterId
	//	}
	//}
	//fmt.Println("Alice is NOT in the cluster: " + clusterId + " She is NOT a receiver.")
	return users, clusterId, ClS

}
