package blog

import (
	"crypto/rand"
	"encoding/hex"
	//"encoding/json"
	"io"
	"os"
	//"time"
	"strings"
)

//var config map[string]interface{}

func GetConfig() map[string]interface{} {
	config := make(map[string]interface{})
	config["author"] = "wendyeq"
	config["title"] = "Wendyeq"
	//config["archive"] = time.Now().Format("2006/01")
	config["archive"] = "2012/11"
	config["themes"] = "bootstrap"
	config["keywords"] = "wendyeq, wendyeq.me, golang, gae, google app engine, mongodb"
	return config
}

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
