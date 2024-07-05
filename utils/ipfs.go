package utils

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"strings"
)

var ipfs *shell.Shell

func GetIPFSClient() *shell.Shell {
	if ipfs == nil {
		ipfs = shell.NewShell("localhost:5001")
	}
	return ipfs
}
func IPFSUpload(msg string) string {
	sh := GetIPFSClient()
	cid, _ := sh.Add(strings.NewReader(msg))
	fmt.Println("Mail IPFS link:", cid)
	return cid
}
