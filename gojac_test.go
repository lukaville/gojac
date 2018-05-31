package gojac

import (
	"testing"
	"fmt"
	"time"
	"io/ioutil"
	"reflect"
)

func TestReadWrite(t *testing.T) {
	sessionInfo := SessionInfo{
		Id:    "session_name",
		Start: time.Now(),
		Dump:  time.Now(),
	}
	executionEntry := ExecutionEntry{
		Id:     12,
		Name:   "execution_entry",
		Probes: []bool{true, false, true},
	}
	executionData := ExecutionData{
		Version:  0x1007,
		Entries:  []ExecutionEntry{executionEntry},
		Sessions: []SessionInfo{sessionInfo},
	}

	tempFile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fail()
	}

	err = Write(tempFile.Name(), executionData)
	if err != nil {
		t.Fail()
	}

	loaded, err := Load(tempFile.Name())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	println(loaded)
	reflect.DeepEqual(executionData, loaded)
}
