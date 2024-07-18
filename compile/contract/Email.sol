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


	function splitAt(string memory _str) pure internal returns (string[] memory){
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
		return res;
	}
	
	mapping(string => PK) public psid2PK;
	mapping(string => mapping(uint64 => string[])) public psid2Day2Cid;
	mapping(string => Mail) public cid2Mail;
	mapping(string => uint256) public cid2Money;
	mapping(string => BcstHeader) public cid2BcstMails;
	mapping(string => Domain) public dmId2Domain;
	mapping(string => uint32[]) public clsId2S;
	mapping(string => DomainId[]) public psid2DmIds;
	mapping(string => string[]) public dm2ClsIds;
	mapping(string => string[]) public psid2TmpPsid;
	mapping(string => mapping(uint64 => string[])) public clsId2Day2Cid;

	struct G1Point {
		uint X; // x-coordinate of point in bn128 G1
		uint Y; // y-coordinate of point in bn128 G1
	}
	struct G2Point {
		uint[2] X; // x-coordinate of point in bn128 G2
		uint[2] Y; // y-coordinate of point in bn128 G2
	}
	struct PK {
		G1Point A;// used in stealth address generation, A= g^a
		G1Point B;// used in stealth address generation, B= g^b
		uint256 fee;// requested minimal fee when receiving an email
        // An address used in receiving digital currency and it does not link to (a,b)
		address payable wallet;
		G1Point[] extra;// stores the stealth address information when the PK is created by others
	}
	struct StealthPub{
		G1Point R; // stealth address used for verification, R =g^r
		G1Point S; // stealth address and the private key s = a+ H(R^b)
	}
	struct ElGamalCT{
		G1Point C1; // the first part of ElGamal ciphertext
		G1Point C2; // the second part of ElGamal ciphertext
	}
	struct Mail{
		StealthPub pub;	// the receiver's stealth address
		ElGamalCT ct;  //ElGamal-encrypted random key {g_i^key}
	}

    struct Domain {
		G1Point[] pArr; // (g,g_1,...,g_n, g_{n+2},...,g_{2n}) and g is bn128 G1 generator 
		G2Point[] qArr; // (h,h_1,...,h_n, h_{n+2},...,h_{2n}) and h is bn128 G2 generator 
		G1Point v; // g^\gamma
		ElGamalCT[] privC;// ElGamal-encrypted private keys {g_i^\gamma}
		string[] psids; // the pseudonyms of each member in the domain
		address admin; // creator of the domain
	}
	struct DomainId {
		uint index; // the index in the domain
		string dmId;// the domain id
	}
	struct BcstHeader {
		G1Point C0; // C0 of BE header
		G1Point C1; // C1 of BE header
		G2Point C0p;// identical C0, but the base is h of G2
	}
	struct DomainProof{
		G1Point skipows; // ski^s, where ski is BE private key and s is a random value
		G2Point pki; // pki, the ith BE public key
		G1Point vpows;// v^s, with the same exponentiation with skipows
		G1Point vpowsp;// v^{s'}, commitment of vpows in sigma protocol
		uint256 c; // challenge value in sigma protocol
		uint256 hatc; // response value in sigma protocol
	}

	mapping(string => G1Point) public dmId2DomainV;

	bool public pairingRes;
	G1Point public pointRes;

	uint256 constant MIN_FEE = 60000;//about $1, with 1ETH=3000$ and gas price = 5Gwei

    function g1add(G1Point memory p1, G1Point memory p2) view internal returns (G1Point memory r) {
		uint[4] memory input;
		input[0] = p1.X;
		input[1] = p1.Y;
		input[2] = p2.X;
		input[3] = p2.Y;
		bool success;
		assembly {
			success := staticcall(sub(gas(), 2000), 6, input, 0xc0, r, 0x60)
			// Use "invalid" to make gas estimation work
			//switch success case 0 { invalid }
		}
		require(success);
	}

    function g1mul(G1Point memory p, uint s) view internal returns (G1Point memory r) {
		uint[3] memory input;
		input[0] = p.X;
		input[1] = p.Y;
		input[2] = s;
		bool success;
		assembly {
			success := staticcall(sub(gas(), 2000), 7, input, 0x80, r, 0x60)
			// Use "invalid" to make gas estimation work
			//switch success case 0 { invalid }
		}
		require (success);
	}

	function equals(
			G1Point memory a, G1Point memory b			
	) pure internal returns (bool) {		
		return a.X==b.X && a.Y==b.Y;
	}
	function register(string memory psid, PK memory pk, string memory oriPsid) public payable returns (PK memory)  {
		require(psid2PK[psid].A.X == 0, "psid exists.");

		if (psid2PK[psid].A.X == 0) {
			psid2PK[psid].A = pk.A;
			psid2PK[psid].B = pk.B;
			psid2PK[psid].fee = pk.fee;
			psid2PK[psid].wallet = pk.wallet;
			for (uint i = 0; i < pk.extra.length; i++) {
				psid2PK[psid].extra.push(G1Point(pk.extra[i].X,pk.extra[i].Y));
			}
			
		}
		// this is used for tempoarily created psid
		if(bytes(oriPsid).length != 0){
			psid2TmpPsid[oriPsid].push(psid);
		}
		return psid2PK[psid];
	}
	function getPK(string memory psid) public view returns (PK memory) {
		return psid2PK[psid];
	}
	
	event Event(string eventName, address indexed sender, uint256 value, string fid, string[] extra);
    // event Event(string eventName, uint256 gasUsed, string[] memory extra);
	
	function mailTo(Mail memory mail, string memory cid, string[] memory psids) public payable {
		// uint256 gasAtStart = gasleft();
		cid2Mail[cid]=mail;
		uint64 currentTime = uint64(block.timestamp);
		uint64 day = currentTime - (currentTime % 86400);

		for (uint i = 0; i < psids.length; i++) {
			psid2Day2Cid[psids[i]][day].push(cid);	
			address payable wallet = psid2PK[psids[i]].wallet;
			uint256 actualValue = psid2PK[psids[i]].fee;
			if (actualValue < MIN_FEE){
				actualValue = MIN_FEE;
			}
			require(msg.value > actualValue, "Mail fees must be greater than MIN_FEE");     
			wallet.transfer(actualValue);
		}
		// uint256 gasUsed = gasAtStart - gasleft(); // 计算消耗的 gas 量		
		emit Event("mailTo", msg.sender, msg.value, cid,psids);

	}

	function getDailyMail(string memory psid, uint64 day) public view returns (string[] memory, Mail[] memory) {
		string[] memory cids = psid2Day2Cid[psid][day];
		Mail[] memory mails = new Mail[](cids.length);
		for (uint i = 0; i < cids.length; i++) {
			mails[i]=cid2Mail[cids[i]];
		}
		return (cids, mails);
	}

	function getTmpPsid(string memory psid) public view returns (string[] memory) {
		return psid2TmpPsid[psid];
	}

	function regDomain(string memory dmId, G1Point[] memory pArr, G2Point[] memory qArr, G1Point memory v, ElGamalCT[] memory privC, string[] memory psids) public payable {
		// G1Point[] memory privC1, G1Point[] memory privC2
		dmId2Domain[dmId].admin = msg.sender;
		dmId2Domain[dmId].v=G1Point(v.X,v.Y);
		dmId2DomainV[dmId]=G1Point(v.X,v.Y);
		for (uint i = 0; i < qArr.length; i++) {//n+1
			dmId2Domain[dmId].qArr.push(G2Point(qArr[i].X,qArr[i].Y));
		}
		for (uint i = 0; i < pArr.length; i++) {//2n+1
			dmId2Domain[dmId].pArr.push(G1Point(pArr[i].X,pArr[i].Y));
		}
		for (uint i = 0; i < privC.length; i++) {//n
			dmId2Domain[dmId].privC.push(ElGamalCT(G1Point(privC[i].C1.X,privC[i].C1.Y), G1Point(privC[i].C2.X,privC[i].C2.Y)));
			// dmId2Domain[dmId].privC2.push());
			dmId2Domain[dmId].psids.push(psids[i]);
			psid2DmIds[psids[i]].push(DomainId(i+1,dmId));
		}
	}

		
	function regCluster(string memory clsId, uint32[] memory S) public payable {
		string[] memory parts = splitAt(clsId);
		Domain memory dm = dmId2Domain[parts[1]];
		if (dm.admin == msg.sender && dm.pArr.length > 0){//cluster should be built when a dm exists
			clsId2S[clsId]= S;
			dm2ClsIds[parts[1]].push(clsId);
		}
	}
	function getS(string memory clsId) public view returns (uint32[] memory) {
		// string[] memory parts = splitAt(clsId);		
		// return dmId2Domain[parts[1]].pArr.length;
		return clsId2S[clsId];
	}

	function getBrdEncPrivs(string memory dmId,string memory psid) public view returns (uint, ElGamalCT memory) {
		ElGamalCT memory ct;
		uint index;
		for (uint i = 0; i < dmId2Domain[dmId].privC.length; i++) {
			string memory psidBC = dmId2Domain[dmId].psids[i];
			if(keccak256(abi.encodePacked(psidBC)) == keccak256(abi.encodePacked(psid))){
				ct = dmId2Domain[dmId].privC[i];
				index=i;
				break;
			}
		}
		return (index,ct);
	}
	function getBrdPKs(string memory dmId) public view returns (G1Point[] memory,G2Point[] memory, G1Point memory) {
		return (dmId2Domain[dmId].pArr, dmId2Domain[dmId].qArr, dmId2Domain[dmId].v);
	}
	
	function bcstTo(BcstHeader memory hdr, string memory clsId, DomainProof memory pi, string memory cid) public payable returns (bool) {
		string[] memory parts = splitAt(clsId);		
		G1Point memory v = dmId2DomainV[parts[1]];
		
		
		// the fees can be put into a buffer, we comment it when testing the gas consumption 
		// string[] memory psids = dm.psids;	
		// uint n =  psids.length;	
		// for (uint i = 0; i < n; i++) {
		// 	PK memory pk = psid2PK[psids[i]];
		// 	uint256 actualValue = msg.value/n;
		// 	if (actualValue < MIN_FEE){
		// 		actualValue = MIN_FEE;
		// 	}
		// 	require(msg.value >= n*actualValue, "Broadcast fees must be greater than n*MIN_FEE");     
		// 	pk.wallet.transfer(actualValue);
		// }

		G1Point[] memory p1Arr = new G1Point[](2);
		G2Point[] memory p2Arr = new G2Point[](2);
		p1Arr[0] = negate(pi.skipows);
		p1Arr[1] = pi.vpows;
		p2Arr[0] = G2;
		p2Arr[1] = pi.pki;

		// pointRes = g1mul(dm.v, pi.hatc);
		
		if(pairing(p1Arr, p2Arr) && equals(g1add(g1mul(pi.vpows, pi.c), g1mul(v, pi.hatc)), pi.vpowsp)) {
			// pairingRes= true;//cost ~20000 gas	
			uint64 currentTime = uint64(block.timestamp);
			uint64 day = currentTime - (currentTime % 86400);
			cid2BcstMails[cid] = hdr;
			clsId2Day2Cid[clsId][day].push(cid);
			emit Event("bcstTo", msg.sender, msg.value, cid, new string[](0));
			return true;
		}else{
			return false;
		}
		return false;
		
	}
	function getPoint() public view returns (G1Point memory) {		
		return pointRes;
	}


	function getDailyBrdMail(string memory clsId, uint64 day) public view returns (string[] memory, BcstHeader[] memory) {
		string[] memory cids = clsId2Day2Cid[clsId][day];
		BcstHeader[] memory mails = new BcstHeader[](cids.length);
		for (uint i = 0; i < cids.length; i++) {
			mails[i]=cid2BcstMails[cids[i]];
		}
		return (cids, mails);
	}

	
	function getPairingRes() public view returns (bool) {
		return pairingRes;
	}

	

	function getMyDomains(string memory psid) public view returns (DomainId[] memory) {
		return psid2DmIds[psid];
	}

	function getMyClusters(string memory dmId) public view returns (string[] memory) {
		return dm2ClsIds[dmId];
	}





}