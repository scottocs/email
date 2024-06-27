package main

import (
	"context"
	"email/compile/contract"
	"email/utils"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func main() {

	contract_name := "Email"

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	privatekey1 := utils.GetENV("PRIVATE_KEY_1")

	deployTX := utils.Transact(client, privatekey1, big.NewInt(0))

	address, _ := utils.Deploy(client, contract_name, deployTX)

	ctc, err := contract.NewContract(common.HexToAddress(address.Hex()), client)

	auth0 := utils.Transact(client, privatekey1, big.NewInt(0))
	tx3, err := ctc.HashToG1(auth0, "11223")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Onchain HashToG1() result:", tx3)
	receipt3, _ := bind.WaitMined(context.Background(), client, tx3)
	fmt.Printf("HashToG1() Gas used: %d\n", receipt3.GasUsed)

}
