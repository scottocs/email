pragma solidity ^0.8.0;
//pragma experimental ABIEncoderV2;
// import "../contracts/bn128G2.sol";
//import "../contracts/strings.sol";
contract Email {
	uint256 constant FIELD_ORDER = 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47;

	// Number of elements in the field (often called `q`)
	// n = n(u) = 36u^4 + 36u^3 + 18u^2 + 6u + 1
	uint256 constant GEN_ORDER = 0x30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001;

	uint256 constant CURVE_B = 3;

	// a = (p+1) / 4
	uint256 constant CURVE_A = 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52;

	struct G1Point {
		uint X;
		uint Y;
	}

	// Encoding of field elements is: X[0] * z + X[1]
	struct G2Point {
		uint[2] X;
		uint[2] Y;
	}


	// (P+1) / 4
	function A() pure internal returns(uint256) {
		return CURVE_A;
	}

	function P() pure internal returns(uint256) {
		return FIELD_ORDER;
	}

	function N() pure internal returns(uint256) {
		return GEN_ORDER;
	}

	/// return the generator of G1
	function P1() pure internal returns(G1Point memory) {
		return G1Point(1, 2);
	}

    // a - b = c;
    function submod(uint a, uint b) internal pure returns (uint){
        uint a_nn;

        if(a>b) {
            a_nn = a;
        } else {
            a_nn = a+GEN_ORDER;
        }

        return addmod(a_nn - b, 0, GEN_ORDER);
    }


    function expMod(uint256 _base, uint256 _exponent, uint256 _modulus)
        internal view returns (uint256 retval)
    {
        bool success;
        uint256[1] memory output;
        uint[6] memory input;
        input[0] = 0x20;        // baseLen = new(big.Int).SetBytes(getData(input, 0, 32))
        input[1] = 0x20;        // expLen  = new(big.Int).SetBytes(getData(input, 32, 32))
        input[2] = 0x20;        // modLen  = new(big.Int).SetBytes(getData(input, 64, 32))
        input[3] = _base;
        input[4] = _exponent;
        input[5] = _modulus;
        assembly {
            success := staticcall(sub(gas(), 2000), 5, input, 0xc0, output, 0x20)
            // Use "invalid" to make gas estimation work
            //switch success case 0 { invalid }
        }
        require(success);
        return output[0];
    }
	

	

	/// return the generator of G2
	function P2() pure internal returns(G2Point memory) {
		return G2Point(
			[11559732032986387107991004021392285783925812861821192530917403151452391805634,
				10857046999023057135944570762232829481370756359578518086990519993285655852781
			],
			[4082367875863433681332203403145435568316851327593401208105741076214120093531,
				8495653923123431417604973247489272438418190587263600148770280649306958101930
			]
		);
	}

	/// return the sum of two points of G1
	function g1add(G1Point memory p1, G1Point memory p2) view internal returns(G1Point memory r) {
		uint[4] memory input;
		input[0] = p1.X;
		input[1] = p1.Y;
		input[2] = p2.X;
		input[3] = p2.Y;
		bool success;
		assembly {
			success:= staticcall(sub(gas(), 2000), 6, input, 0xc0, r, 0x60)
			// Use "invalid" to make gas estimation work
			//switch success case 0 { invalid }
		}
		require(success);
	}

	/// return the product of a point on G1 and a scalar, i.e.
	/// p == p.mul(1) and p.add(p) == p.mul(2) for all points p.
	function g1mul(G1Point memory p, uint s) view internal returns(G1Point memory r) {
		uint[3] memory input;
		input[0] = p.X;
		input[1] = p.Y;
		input[2] = s;
		bool success;
		assembly {
			success:= staticcall(sub(gas(), 2000), 7, input, 0x80, r, 0x60)
			// Use "invalid" to make gas estimation work
			//switch success case 0 { invalid }
		}
		require(success);
	}

	/// return the result of computing the pairing check
	/// e(p1[0], p2[0]) *  .... * e(p1[n], p2[n]) == 1
	/// For example pairing([P1(), P1().negate()], [P2(), P2()]) should
	/// return true.
	function pairing(G1Point[] memory p1, G2Point[] memory p2) view internal returns(bool) {
		require(p1.length == p2.length);
		uint elements = p1.length;
		uint inputSize = elements * 6;
		uint[] memory input = new uint[](inputSize);
		for (uint i = 0; i < elements; i++) {
			input[i * 6 + 0] = p1[i].X;
			input[i * 6 + 1] = p1[i].Y;
			input[i * 6 + 2] = p2[i].X[0];
			input[i * 6 + 3] = p2[i].X[1];
			input[i * 6 + 4] = p2[i].Y[0];
			input[i * 6 + 5] = p2[i].Y[1];
		}
		uint[1] memory out;
		bool success;
		assembly {
			success:= staticcall(sub(gas(), 2000), 8, add(input, 0x20), mul(inputSize, 0x20), out, 0x20)
			// Use "invalid" to make gas estimation work
			//switch success case 0 { invalid }
		}
		require(success);
		return out[0] != 0;
	}


	function equals(
		G1Point memory a, G1Point memory b
	) view internal returns(bool) {
		return a.X == b.X && a.Y == b.Y;
	}

	function equals2(
		G2Point memory a, G2Point memory b
	) view internal returns(bool) {
		return a.X[0] == b.X[0] && a.X[1] == b.X[1] && a.Y[0] == b.Y[0] && a.Y[1] == b.Y[1];
	}

	function HashToG1(string memory str) public payable returns(G1Point memory) {

		return g1mul(P1(), uint256(keccak256(abi.encodePacked(str))));
	}

	function negate(G1Point memory p) public payable returns(G1Point memory) {
		// The prime q in the base field F_q for G1
		uint q = 21888242871839275222246405745257275088696311157297823662689037894645226208583;
		if (p.X == 0 && p.Y == 0)
			return G1Point(0, 0);
		return G1Point(p.X, q - (p.Y % q));
	}

	function stringEqual(
		string memory a,
		string memory b
	) private pure returns(bool same) {
		return keccak256(bytes(a)) == keccak256(bytes(b));
	}

	function bytesEqual(
		bytes memory a,
		bytes memory b
	) private pure returns(bool same) {
		return keccak256(a) == keccak256(b);
	}
	mapping(string => PK) public name2PK;
	mapping(string => MailRev) public cid2Mails;
	mapping(string => BrdcastMailRev) public cid2BrdcastMails;

	struct PK {
		G1Point A;
		G1Point B;
	}
	struct StealthAddrPub{
		G1Point R;
		G1Point S;
	}
	struct BrdcastCT{
		G1Point C0;
		G1Point C1;
	}
	struct ElGamalCT{
		G1Point C1;
		G1Point C2;
	}
	struct DomainProof{
		G1Point skipows;
		G1Point pki;
		G1Point vpows;
	}
	struct MailRev{
		StealthAddrPub pub;		
		ElGamalCT ct;
		string[] names;
	}

	struct BrdcastMailRev{
		BrdcastCT ct;
		DomainProof proof;
		string[] names;
	}

	function uploadPK(string memory name, PK memory pk) public payable returns (PK memory)  {
		require(name2PK[name].A.X == 0, "name does not exist.");

		if (name2PK[name].A.X == 0) {
			name2PK[name] = pk;
		}
		return name2PK[name];
	}
	function downloadPK(string memory name) public view returns (PK memory) {
		return name2PK[name];
	}

	function mailTo(StealthAddrPub memory saPub, ElGamalCT memory ct, string memory cid, string[] memory names) public payable  {
		cid2Mails[cid] = MailRev(saPub, ct, names);
		//TODO	emit event
	}
	function brdcastTo(BrdcastCT memory ct, DomainProof memory proof, string memory cid, string[] memory names) public payable  {
		cid2BrdcastMails[cid] = BrdcastMailRev(ct, proof, names);
		//TODO	emit event
	}
	function downloadMail(string memory cid) public view returns (MailRev memory) {
		return cid2Mails[cid];
	}
}