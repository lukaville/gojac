package gojac

import (
	"testing"
	"fmt"
	"time"
	"reflect"
	"io/ioutil"
)

func TestReadMinimalFile(t *testing.T) {
	loaded, err := Read("fixtures/simple1.exec")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	if len(loaded.Entries) != 0 {
		t.Fail()
	}
}

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
		fmt.Println(err)
		t.Fail()
	}

	err = Write(tempFile.Name(), executionData)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	loaded, err := Read(tempFile.Name())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	
	reflect.DeepEqual(executionData, loaded)
}
