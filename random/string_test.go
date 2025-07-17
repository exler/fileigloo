package random_test

import (
	"strings"
	"testing"

	"github.com/exler/fileigloo/random"
)

func TestRandomString(t *testing.T) {
	str := random.String(6)

	if len(str) != 6 {
		t.Errorf("len(String(6)) != 6")
	}

	for _, char := range str {
		if !strings.Contains("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", string(char)) {
			t.Errorf("Byte outside the alphabet used in random string")
		}
	}
}
