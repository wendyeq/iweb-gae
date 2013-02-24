// +build appengine
package blog

import (
	"crypto/rand"
	"encoding/hex"
	//"encoding/json"
	"io"
	"os"
	"strings"
)

//from www.ashishbanerjee.com/home/go/go-generate-uuid
func GenUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = 0x80
	uuid[4] = 0x40
	return hex.EncodeToString(uuid), nil
}

//var ctx Context
func GetContext() Context {
	ctx := Context{}
	args := make(map[string]interface{})
	args["author"] = "wendyeq"
	args["title"] = "Wendyeq"
	args["brandName"] = "Wendyeq"
	args["archive"] = "2012/11"
	args["themes"] = "bootstrap"
	args["keywords"] = "wendyeq, wendyeq.me, Go, golang, gae, google app engine, mongodb"
	args["size"] = 5
	args["pageSize"] = 1
	ctx.Args = args
	return ctx
}

func GetRelease() (buf []byte, err error) {
	file, err := os.Open("RELEASE.md")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf = make([]byte, 10240)
	io.ReadFull(file, buf)

	return buf, err
}

func (str NewString) Replace(oldStr, newStr string, n int) string {
	return strings.Replace(string(str), oldStr, newStr, n)
}
