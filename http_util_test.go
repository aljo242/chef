package http_util

import "testing"

func TestBasic(t *testing.T) {
	// t.Fatal("not implemented")
	err := Print()
	if err != nil {
		t.Errorf("OOPS : %w", err)
	}
}
