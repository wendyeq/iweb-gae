// +build !appengine
package blog

import (
	//appenginetesting "github.com/tenntenn/gae-go-testing"
	"syscall"
	"testing"
)

func TestGenUUID(t *testing.T) {
	uuid, err := GenUUID()
	if err != nil {
		t.Fatalf("Generate UUID Error : %v", err)
	}
	if len(uuid) != 32 {
		t.Fatalf("uuid len is not 32, the Len is : %v, the uuid is : %v", len(uuid), uuid)
	}
	t.Log(syscall.AF_ALG)
}

func BenchmarkGenUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenUUID()
	}
}

func TestGetConfig(t *testing.T) {
	m := GetConfig()
	if len(m) <= 0 {
		t.Fatalf("GetConfig err: %v", len(m))
	}
	if _, ok := m["author"]; !ok {
		t.Fatal("GetConfig author is not exist!")
	}
}

func BenchmarkGetConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetConfig()
	}
}

func TestGetRelease(t *testing.T) {
	buf, err := GetRelease()
	if err != nil {
		t.Fatalf("GetRelease error. The err is : %v", err)
	}
	if len(buf) <= 0 {
		t.Fatalf("GetRelease failed. Didn't have release content.")
	}
}

func BenchmarkGetRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetRelease()
	}
}
