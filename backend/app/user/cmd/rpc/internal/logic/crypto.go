package logic

import (
	"crypto/md5"
	"encoding/hex"
)

func md5ByString(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
