# Proof of concept implementation for "Email in the blockchain era"

# Pre-requisites

* `Golang`  https://go.dev/dl/   

* `Solidity`  https://docs.soliditylang.org/en/v0.8.2/installing-solidity.html  Version: 0.8.20

* `Solidity compiler (solc)`  https://docs.soliditylang.org/en/latest/installing-solidity.html  
Version: 0.8.25-develop

* `Ganache-cli`  https://www.npmjs.com/package/ganache-cli
    
* `Abigen`    Version: v1.14.3
    ```bash
    go get -u github.com/ethereum/go-ethereum
    go install github.com/ethereum/go-ethereum/cmd/abigen@v1.14.3
    ```


# File description

* `tests/*`   test the functionalities of the framework.

* `compile/contract/`  The folder stores contract source code file (.sol) and generated go contract file.

* `compile/compile.sh`  The script file compiles solidity and generates go contract file.

* `genPrvKey.sh`  The script file generates accounts and stores in the`.env` file.


# How to run

1. Generate private keys to generate the `.env` file. Be sure that ganache is not started when runing below command.

    ```bash
    bash genPrvKey.sh
    ```

2. start ganache

    ```bash
    ganache --mnemonic "email" -l 90071992547 -e 100000
    ```

3. Compile the smart contract code

    ```bash
    bash compile.sh
    ```
4. Start the IPFS serivce
5. Start the monitor process (optional)

    ```bash
    python monitorEvent.py
    ```
6. Test the Email system
    ```bash
    cd tests
    go test
    ```
    or 
    ```bash
    go test -v -run TestBcstLinkableCluster
    ```
    or
    ```bash
    go test Bcst_test.go -v
    ```
