package utils

import (
	"email/crypto/broadcast"
	"github.com/fentec-project/bn256"
	"math/big"
)

type User struct {
	Name       string
	Aa         *big.Int    //Stealth address private key a
	Bb         *big.Int    //Stealth address private key b
	A          *bn256.G1   //Stealth address public key A (bn256)
	B          *bn256.G1   //Stealth address public key B (bn256)
	Privatekey string      // ethereum private key (secp256k1)
	BrdUser    *DomainUser //public keys (bn256) used in broadcast email
}
type Group struct {
	PK broadcast.CompletePublicKey
	SK broadcast.AdvertiserSecretKey
}
type DomainUser struct { // A group user is a domain user
	Group    Group
	S        []int
	DomainPK *bn256.G1
}
