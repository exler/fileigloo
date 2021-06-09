package random

import (
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	str := String(6)

	if len(str) != 6 {
		t.Errorf("len(String(6)) != 6")
	}

	for _, char := range str {
		if !strings.Contains(letterBytes, string(char)) {
			t.Errorf("Byte outside the alphabet used in random string")
		}
	}
}
