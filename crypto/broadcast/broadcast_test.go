package broadcast

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func convertToInt32(arr []int) []uint32 {
	result := make([]uint32, len(arr))
	for i, v := range arr {
		result[i] = uint32(v)
	}
	return result
}
func randomPick(n, t int) []int {
	rand.Seed(time.Now().UnixNano()) // 设置随机种子
	indices := rand.Perm(n)[:t]      // 生成随机排列并取前 t 个
	result := make([]int, t)
	for i, idx := range indices {
		result[i] = idx + 1 // 转换为 1-n 范围
	}
	return result
}

func TestBroadcast(te *testing.T) {
	n := 100
	t := n/2 + 1
	//t := 2*n/3 + 1

	cpk, secretKeys := Setup(n, "dmId")
	//fmt.Println(cpk)
	//fmt.Println(secretKeys)

	//S := convertToInt32([]int{1, 3})
	S := convertToInt32(randomPick(n, t))
	fmt.Printf("S %v\n", S)
	//clusterPK := cpk.buildClusterPK(S)
	var times int64 = 1000
	starttime := time.Now().UnixMicro()
	for i := 0; i < int(times); i++ {
		cpk.Encrypt(S)
	}
	endtime := time.Now().UnixMicro()
	fmt.Printf("Encrypt time cost %d us\n", (endtime-starttime)/times)
	hdr, beK := cpk.Encrypt(S)
	//hdr, K, err := bpk.Encrypt(S)

	beKp := secretKeys[S[1]].Decrypt(S, hdr, cpk)
	starttime = time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		beKp = secretKeys[S[1]].Decrypt(S, hdr, cpk)
	}
	endtime = time.Now().UnixMicro()
	fmt.Printf("Decrypt time cost %d us\n", (endtime-starttime)/times)
	if beK.String() != beKp.String() {
		fmt.Printf("Equality check failed\nK: %v\nchkK: %v", beK.String(), beKp.String())
	}
	fmt.Printf("Equality check success\n")
}
