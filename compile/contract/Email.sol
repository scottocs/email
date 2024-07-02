// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;


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
	G1Point G1 = G1Point(1, 2);
    G2Point G2 = G2Point(
        [11559732032986387107991004021392285783925812861821192530917403151452391805634,
        10857046999023057135944570762232829481370756359578518086990519993285655852781],
        [4082367875863433681332203403145435568316851327593401208105741076214120093531,
        8495653923123431417604973247489272438418190587263600148770280649306958101930]
    );


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
	mapping(string => PK) public psid2PK;
	mapping(string => mapping(uint64 => string[])) public psid2Day2Cid;
	mapping(string => MailRev) public cid2Mail;
	mapping(string => BrdcastHeader) public cid2BrdcMails;
	mapping(string => Domain) public domainId2Domain;
	mapping(string => uint32[]) public clusterId2S;
	mapping(string => string[] ) public psid2GrpIds;
	mapping(string => mapping(uint64 => string[])) public clusterId2Day2Cid;

	bool public pairingRes;

	struct PK {
		G1Point A;
		G1Point B;
	}
	struct StealthAddrPub{
		G1Point R;
		G1Point S;
	}
	struct Domain {
		G1Point[] pArr;
		G2Point[] qArr;
		G1Point v;
		G1Point[] privC1;
		G1Point[] privC2;
		string[] psids;
	}

	struct BrdcastHeader {
		G1Point C0;
		G1Point C1;
		G2Point C0p;
	}
	struct ElGamalCT{
		G1Point C1;
		G1Point C2;
	}
	struct ClusterProof{
		G1Point skipows;
		G2Point pki;
		G1Point vpows;
	}
	struct MailRev{
		StealthAddrPub pub;		
		ElGamalCT ct;
	}

	function register(string memory name, PK memory pk) public payable returns (PK memory)  {
		require(psid2PK[name].A.X == 0, "name exists.");

		if (psid2PK[name].A.X == 0) {
			psid2PK[name] = pk;
		}
		return psid2PK[name];
	}
	function downloadPK(string memory name) public view returns (PK memory) {
		return psid2PK[name];
	}
//	uint64 today;
	function mailTo(StealthAddrPub memory saPub, ElGamalCT memory ct, string memory cid, string[] memory psids) public payable {
		cid2Mail[cid]=MailRev(saPub, ct);
		uint64 currentTime = uint64(block.timestamp);
		uint64 day = currentTime - (currentTime % 86400);

		for (uint i = 0; i < psids.length; i++) {
			psid2Day2Cid[psids[i]][day].push(cid);			
		}
		//TODO	emit event
	}

	function getDailyMail(string memory psid, uint64 day) public view returns (string[] memory, MailRev[] memory) {
		string[] memory cids = psid2Day2Cid[psid][day];
		MailRev[] memory mails = new MailRev[](cids.length);
		for (uint i = 0; i < cids.length; i++) {
			mails[i]=cid2Mail[cids[i]];
		}
		return (cids, mails);
	}

	function regDomain(string memory domainId, G1Point[] memory pArr, G2Point[] memory qArr, G1Point memory v,G1Point[] memory privC1, G1Point[] memory privC2, string[] memory psids) public payable {
//		DomainParams storage domain = ;
		domainId2Domain[domainId].v=G1Point(v.X,v.Y);
		for (uint i = 0; i < qArr.length; i++) {//n+1
			domainId2Domain[domainId].qArr.push(G2Point(qArr[i].X,qArr[i].Y));
		}
		for (uint i = 0; i < pArr.length; i++) {//2n+1
			domainId2Domain[domainId].pArr.push(G1Point(pArr[i].X,pArr[i].Y));
		}
		for (uint i = 0; i < privC1.length; i++) {//n
			domainId2Domain[domainId].privC1.push(G1Point(privC1[i].X,privC1[i].Y));
			domainId2Domain[domainId].privC2.push(G1Point(privC2[i].X,privC2[i].Y));
			domainId2Domain[domainId].psids.push(psids[i]);
			psid2GrpIds[psids[i]].push(domainId);
		}
	}

	string[] public str;
	function splitAt(string memory _str) public view returns (string[] memory){
		bytes memory sbt = bytes(_str);
		string[] memory res = new string[](2);
		uint len = 0;
		for (uint i = 0; i < sbt.length; i++) {
			if(bytes1('@')==sbt[i]){
				len = i;
				break;
			}
		}
		bytes memory left = new bytes(len);
		bytes memory right = new bytes(sbt.length - len-1);
		for (uint i = 0; i < sbt.length; i++) {
			if( i< len){
				left[i] = sbt[i];
			}
			if (i> len){
				right[i-len-1] = sbt[i];
			}
		}
		res[0]=string(left);
		res[1]=string(right);
//		str.push(res[0]);
//		str.push(res[1]);
		return res;
	}
	// function downloadSplit(string memory clusterId) public view returns (string[] memory) {
	// 	return str;
	// }
	function regCluster(string memory clusterId, uint32[] memory S) public payable {
		string[] memory parts = splitAt(clusterId);
		Domain memory domain = domainId2Domain[parts[1]];
		if (domain.pArr.length > 0){//cluster should be built when a domain exists
			clusterId2S[clusterId]= S;
		}

	}
	function getS(string memory clusterId) public view returns (uint32[] memory) {
		string[] memory parts = splitAt(clusterId);		
		// return domainId2Domain[parts[1]].pArr.length;
		return clusterId2S[clusterId];
	}

	function retrBrdPrivs(string memory domainId,string memory name) public view returns (uint, G1Point memory,G1Point memory) {
		G1Point memory c1;
		G1Point memory c2;
		uint index;
		for (uint i = 0; i < domainId2Domain[domainId].privC1.length; i++) {
			string memory nameBC = domainId2Domain[domainId].psids[i];
			if(keccak256(abi.encodePacked(nameBC)) == keccak256(abi.encodePacked(name))){
				c1= domainId2Domain[domainId].privC1[i];
				c2= domainId2Domain[domainId].privC2[i];
				index=i;
				break;
			}
		}
		return (index,c1, c2);
	}
	function getBrdPKs(string memory domainId) public view returns (G1Point[] memory,G2Point[] memory, G1Point memory) {
		return (domainId2Domain[domainId].pArr, domainId2Domain[domainId].qArr, domainId2Domain[domainId].v);
	}
	// function DownloadClusterPK(string memory domainId) public view returns (G1Point[] memory,G2Point[] memory, G1Point memory) {
	// 	return (domainId2Domain[domainId].pArr, domainId2Domain[domainId].qArr, domainId2Domain[domainId].v);
	// }

	function bcstTo(BrdcastHeader memory hdr, string memory clusterId, ClusterProof memory proof, string memory cid) public payable returns (bool)  {
		// todo anonymoty of senders
		G1Point[] memory p1Arr = new G1Point[](2);
		G2Point[] memory p2Arr = new G2Point[](2);
		p1Arr[0] = negate(proof.skipows);
		p1Arr[1] = proof.vpows;
		p2Arr[0] = G2;
		p2Arr[1] = proof.pki;

		if(pairing(p1Arr, p2Arr)) {
			// pairingRes= true;//cost ~20000 gas	
			uint64 currentTime = uint64(block.timestamp);
			uint64 day = currentTime - (currentTime % 86400);
			cid2BrdcMails[cid] = hdr;
			clusterId2Day2Cid[clusterId][day].push(cid);
			return true;
		}else{
			return false;
		}
		//TODO	emit event
	}

	function getDailyBrdMail(string memory clusterId, uint64 day) public view returns (string[] memory, BrdcastHeader[] memory) {
		string[] memory cids = clusterId2Day2Cid[clusterId][day];
		BrdcastHeader[] memory mails = new BrdcastHeader[](cids.length);
		for (uint i = 0; i < cids.length; i++) {
			mails[i]=cid2BrdcMails[cids[i]];
		}
		return (cids, mails);
	}

	
	function getPairingRes() public view returns (bool) {
		return pairingRes;
	}

	

	function getMyDomains(string memory psid) public view returns (string[] memory) {
		return psid2GrpIds[psid];
	}





}