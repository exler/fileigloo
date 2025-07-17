package datetime_test

import (
	"testing"

	"github.com/exler/fileigloo/datetime"
)

func TestStringExpired(t *testing.T) {
	str := "2023-10-01T00:00:00Z"
	if !datetime.IsExpired(str) {
		t.Errorf("Expected %s to be expired", str)
	}
}

func TestStringNotExpired(t *testing.T) {
	str := "2099-10-01T12:00:00Z"
	if datetime.IsExpired(str) {
		t.Errorf("Expected %s to not be expired", str)
	}
}

func TestStringEmpty(t *testing.T) {
	str := ""
	if datetime.IsExpired(str) {
		t.Errorf("Expected empty string to not be expired")
	}
}

func TestStringInvalid(t *testing.T) {
	str := "invalid-date"
	if datetime.IsExpired(str) {
		t.Errorf("Expected invalid date string to not be expired")
	}
}
