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
	struct G1Point {
		uint X; // x-coordinate of point in bn128 G1
		uint Y; // y-coordinate of point in bn128 G1
	}
	struct G2Point {
		uint[2] X; // x-coordinate of point in bn128 G2
		uint[2] Y; // y-coordinate of point in bn128 G2
	}
	
	mapping(string => PK) psid2PK;
	mapping(string => mapping(uint64 => string[])) psid2Day2Cid;
	mapping(string => Mail) cid2SA;
	

	//seperately store fields in dmId2Domain/dmId2PArr/dmId2QArr/dmId2Psids/StealthEncPriv/psid2DmIds can save gas cost when accessing them
	mapping(string => Domain) dmId2Domain;
	mapping(string => G1Point[]) dmId2PArr; // (g,g_1,...,g_n, g_{n+2},...,g_{2n}) on G1
	mapping(string => G2Point[]) dmId2QArr; // (h,h_1,...,h_n, h_{n+2},...,h_{2n}) on G2
	mapping(string => string[]) dmId2Psids; // the pseudonyms of each member in the domain
	mapping(string => mapping(string => StealthEncPriv)) dmId2Psid2PrivC;// ElGamal-encrypted private keys {g_i^\gamma}	
	mapping(string => StealthPub[]) dmId2SAPubs;
	mapping(uint => bool) SAPubS2SAPubR;
	mapping(uint256 => bool) Pi2Used;
	mapping(string => DomainId[]) psid2DmIds;

	mapping(string => string[]) dmId2ClsIds;
	mapping(string => EncClS) clsId2EncS;	

	
	mapping(string => BcstHeader) cid2BcstHdr;	
	mapping(string => mapping(uint64 => string[])) clsId2Day2Cid;

	
	struct PK {
		G1Point A;// used in stealth address generation, A= g^a
		G1Point B;// used in stealth address generation, B= g^b
		uint256 fee;// requested minimal fee when receiving an email        
		address payable wallet; // An address used in receiving digital currency
		G1Point[] extra; // stores the stealth address for temporary a user
	}

	struct StealthPub{
		G1Point R; // stealth address used for verification, R =g^r
		G1Point S; // stealth address and the private key s = a+ H(R^b)
	}
	
	// encrypted using a stealth address
	struct StealthEncPriv{
		G1Point C1; // the first part of ElGamal ciphertext
		G1Point C2; // the second part of ElGamal ciphertext
	}

	struct Mail{
		StealthPub pub;	// the receiver's stealth address
		G1Point C1; // the first part of ElGamal ciphertext
		G1Point C2; // the second part of ElGamal ciphertext
	}

    struct Domain {
		G1Point v; // g^\gamma
		address payable admin; // creator of the domain
		uint256 fee;
		uint256 deposits;
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

	struct EncClS {
		BcstHeader hdr; // BE ciphertext
		string str; // ciphertext of ClS
	}
	struct Pi{
		G1Point Cp;// g^{s'}, Schnorr sigma protocol commitment
		uint256 c; // Sigma protocol challenge value 
		uint256 ctilde; // Sigma protocol response
	}

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
	function register(string memory psid, PK memory pk) public payable returns (PK memory)  {
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
		return psid2PK[psid];
	}
	function getPK(string memory psid) public view returns (PK memory) {
		return psid2PK[psid];
	}
	
	event Event(string eventName, address indexed sender, uint256 value, string fid, string[] extra);
    // event Event(string eventName, uint256 gasUsed, string[] memory extra);
	
	function oto(Mail memory mail, string memory cid, string[] memory psids) public payable {
		// uint256 gasAtStart = gasleft();
		cid2SA[cid]=mail;
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
			// emit Event("oto", wallet, actualValue, cid,psids);
		}
		// uint256 gasUsed = gasAtStart - gasleft(); // 计算消耗的 gas 量		
		emit Event("oto", msg.sender, msg.value, cid,psids);

	}

	function getDailyMail(string memory psid, uint64 day) public view returns (string[] memory, Mail[] memory) {
		string[] memory cids = psid2Day2Cid[psid][day];
		Mail[] memory mails = new Mail[](cids.length);
		for (uint i = 0; i < cids.length; i++) {
			mails[i]=cid2SA[cids[i]];
		}
		return (cids, mails);
	}

	
	function regDomain(string memory dmId, G1Point[] memory pArr, G2Point[] memory qArr, G1Point memory v, StealthEncPriv[] memory encPriv, string[] memory psids, StealthPub[] memory saPubs) public payable {
		dmId2Domain[dmId].admin = payable(msg.sender);
		dmId2Domain[dmId].fee = 0;
		dmId2Domain[dmId].v=v;
		for (uint i = 0; i < pArr.length; i++) {//2n+1
			dmId2PArr[dmId].push(pArr[i]);
		}
		for (uint i = 0; i < qArr.length; i++) {//n+1
			dmId2QArr[dmId].push(qArr[i]);
		}
		for (uint i = 0; i < encPriv.length; i++) {//n
			// note that saPubs and psids are not linked
			dmId2Psid2PrivC[dmId][psids[i]]=encPriv[i];
			dmId2SAPubs[dmId].push(saPubs[i]);
			SAPubS2SAPubR[saPubs[i].S.X] = true;// to enable quick query
			dmId2Psids[dmId].push(psids[i]);
			psid2DmIds[psids[i]].push(DomainId(i+1,dmId));
			PK memory pk = psid2PK[psids[i]];
			dmId2Domain[dmId].fee+=pk.fee;
			dmId2Domain[dmId].deposits=0;
		}
	}

		
	function regCluster(string memory clsId, string memory ClSEncStr, BcstHeader memory hdr) public payable {
		string[] memory parts = splitAt(clsId);
		Domain memory dm = dmId2Domain[parts[1]];
		if (dm.admin == msg.sender){//cluster should be built when a dm exists
			clsId2EncS[clsId]= EncClS(hdr, ClSEncStr);
			dmId2ClsIds[parts[1]].push(clsId);
		}
	}
	function getEncClS(string memory clsId) public view returns (EncClS memory) {
		return clsId2EncS[clsId];
	}

	function getBrdEncPrivs(string memory dmId, string memory psid) public view returns (StealthEncPriv memory, StealthPub[] memory) {
		return (dmId2Psid2PrivC[dmId][psid], dmId2SAPubs[dmId]);
	}
	function getBrdPKs(string memory dmId) public view returns (G1Point[] memory,G2Point[] memory, G1Point memory) {
		return (dmId2PArr[dmId], dmId2QArr[dmId], dmId2Domain[dmId].v);
	}
	
	receive() external payable {
        emit Event("contract_receive", msg.sender, msg.value, "", new string[](0));
    }
	function bcstTo(BcstHeader memory hdr, string memory clsId, Pi memory pi, string memory cid, G1Point memory SAPubS) public payable returns (bool) {
		string[] memory parts = splitAt(clsId);
		require(msg.value > dmId2Domain[parts[1]].fee, "Broadcast fees must be greater than required");     
		dmId2Domain[parts[1]].deposits += msg.value;
		// msg.value is automatically tranferred to the contract

		if(Pi2Used[pi.c]!=true && SAPubS2SAPubR[SAPubS.X] && equals(g1add(g1mul(SAPubS, pi.c), g1mul(P1(), pi.ctilde)), pi.Cp)) {
			// pairingRes= true;//cost ~20000 gas	
			uint64 currentTime = uint64(block.timestamp);
			uint64 day = currentTime - (currentTime % 86400);
			cid2BcstHdr[cid] = hdr;
			clsId2Day2Cid[clsId][day].push(cid);
			string[] memory res = new string[](1);
			res[0]=string(clsId);
			Pi2Used[pi.c]=true;
			emit Event("bcstTo", msg.sender, msg.value, cid, res);
			return true;
		}else{
			return false;
		}
		// return false;		
	}
	function reward(string memory dmId)  public payable {
		// // the fees can be put into a deposit (buffer), we comment it when testing the gas consumption 				
		string[] memory psids = dmId2Psids[dmId];		
		uint256 totalFee = dmId2Domain[dmId].deposits;
		
		if (totalFee<=0){
			return;
		}
		
		uint n =  psids.length;	
		uint256 total = 0;
		for (uint i = 0; i < n; i++) {
			PK memory pk = psid2PK[psids[i]];
			total+= pk.fee;
		}
		

		dmId2Domain[dmId].deposits=0;
		for (uint i = 0; i < n; i++) {
			PK memory pk = psid2PK[psids[i]];
			uint256 val = totalFee/total * pk.fee;
			address payable wallet = psid2PK[psids[i]].wallet;
			// address payable wallet = pk.wallet;
			wallet.transfer(val);			
		}
		emit Event("withdraw", dmId2Domain[dmId].admin, address(this).balance, dmId, new string[](0));
	}
	function getPoint() public view returns (G1Point memory) {		
		return pointRes;
	}


	function getDailyBrdMail(string memory clsId, uint64 day) public view returns (string[] memory, BcstHeader[] memory) {
		string[] memory cids = clsId2Day2Cid[clsId][day];
		BcstHeader[] memory hdrs = new BcstHeader[](cids.length);
		for (uint i = 0; i < cids.length; i++) {
			hdrs[i]=cid2BcstHdr[cids[i]];
		}
		return (cids, hdrs);
	}

	
	function getPairingRes() public view returns (bool) {
		return pairingRes;
	}

	

	function getMyDomains(string memory psid) public view returns (DomainId[] memory) {
		return psid2DmIds[psid];
	}

	function getMyClusters(string memory dmId) public view returns (string[] memory) {
		return dmId2ClsIds[dmId];
	}





}