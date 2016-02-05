package main

import (
	"testing"
)

func TestQueryToken(t *testing.T) {
	token, err := queryToken("Github_Token", "vampirewalk")
	if err != nil {
		t.Errorf("Token Not Found")
	}
}
