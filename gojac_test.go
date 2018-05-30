package gojac

import (
	"testing"
	"fmt"
)

func TestLoad(t *testing.T) {
	data, err := Load("fixtures/simple1.exec")

	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	println(data)
}