package tests

import (
	"fmt"
	"os"
	"testing"
)

func TestExecutable(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(pwd)
	t.Error("")
}
