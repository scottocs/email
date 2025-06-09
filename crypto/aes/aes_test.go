package aes

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateString(length int) string {
	rand.Seed(time.Now().UnixNano()) // Set random seed
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func TestAES(t *testing.T) {
	key := []byte(generateString(32))
	msg := []byte(generateString(1024 * 1000 * 210))
	ct, _ := Encrypt(msg, key)
	var n int64 = 1
	starttime := time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		ct, _ = Encrypt(msg, key)
	}
	endtime := time.Now().UnixMicro()
	fmt.Printf("Encrypt time cost %d us\n", (endtime-starttime)/n)

	pt, _ := Decrypt(ct, key)

	starttime = time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		pt, _ = Decrypt(ct, key)
	}
	endtime = time.Now().UnixMicro()
	fmt.Printf("Decrypt time cost %d us. %v\n", (endtime-starttime)/n, bytes.Equal([]byte(pt), msg))

}
