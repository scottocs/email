package utils

import (
	"email/crypto/broadcast"
	"github.com/fentec-project/bn256"
	"math/big"
)

type User struct {
	Psid       string
	Aa         *big.Int         //Stealth address private key a
	Bb         *big.Int         //Stealth address private key b
	A          *bn256.G1        //Stealth address public key A (bn256)
	B          *bn256.G1        //Stealth address public key B (bn256)
	Privatekey string           // ethereum private key (secp256k1)
	Groups     map[string]Group // used if the user is in a group
}
type Group struct { //
	PKs     broadcast.PKs
	SK      broadcast.SK
	Domains map[string][]uint32
}
