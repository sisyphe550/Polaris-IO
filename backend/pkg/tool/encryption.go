package tool

import (
	"crypto/md5"
	"fmt"
	"io"
)

// Md5ByString 对字符串进行 MD5 加密
func Md5ByString(str string) string {
	m := md5.New()
	_, err := io.WriteString(m, str)
	if err != nil {
		panic(err)
	}
	arr := m.Sum(nil)
	return fmt.Sprintf("%x", arr)
}

// Md5ByBytes 对字节数组进行 MD5 加密
func Md5ByBytes(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}
