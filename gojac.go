package gojac

import (
	"bufio"
	"os"
	"fmt"
	"encoding/binary"
	"errors"
	"time"
)

const (
	headerMagicNumber    uint16 = 0xC0C0
	endOfFileMarker      byte   = 0xFF
	headerMarker         byte   = 0x01
	sessionInfoMarker    byte   = 0x10
	executionEntryMarker byte   = 0x11
)

func Load(path string) (*ExecutionData, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	data := &ExecutionData{}

	for {
		marker, _ := reader.ReadByte()

		switch marker {
		case headerMarker:
			data.Version, err = readHeader(reader)
		case sessionInfoMarker:
			sessionInfo, err := readSessionInfo(reader)
			if err != nil {
				return nil, err
			}
			data.Sessions = append(data.Sessions, *sessionInfo)
		case executionEntryMarker:
			entry, err := readEntry(reader)
			if err != nil {
				return nil, err
			}
			data.Entries = append(data.Entries, *entry)
		case endOfFileMarker:
			return data, nil
		}
	}
}

func readHeader(reader *bufio.Reader) (int16, error) {
	var magicNumber uint16
	err := binary.Read(reader, binary.LittleEndian, &magicNumber)

	if err != nil {
		return 0, err
	}

	if magicNumber != headerMagicNumber {
		return 0, errors.New(fmt.Sprintf("readSessionInfo: invalid magic number %v (expected %v)", magicNumber, headerMagicNumber))
	}

	var fileVersion int16
	err = binary.Read(reader, binary.LittleEndian, &fileVersion)

	if err != nil {
		return 0, err
	}

	return fileVersion, nil
}

func readEntry(reader *bufio.Reader) (*ExecutionEntry, error) {
	entry := &ExecutionEntry{}

	var id int64
	err := binary.Read(reader, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}
	entry.Id = id

	name, err := readString(reader)
	if err != nil {
		return nil, err
	}
	entry.Name = name

	var probesCount int32
	err = binary.Read(reader, binary.LittleEndian, &probesCount)
	if err != nil {
		return nil, err
	}

	probes := make([]bool, probesCount)
	var buffer byte = 0x00
	for i := 0; i < len(probes); i++ {
		if (i % 8) == 0 {
			buffer, err = reader.ReadByte()
			if err != nil {
				return nil, err
			}
		}
		probes[i] = (buffer & 0x01) != 0
		buffer = buffer >> 1
	}
	entry.Probes = probes

	return entry, nil
}

func readSessionInfo(reader *bufio.Reader) (*SessionInfo, error) {
	sessionInfo := &SessionInfo{}

	id, err := readString(reader)
	if err != nil {
		return nil, err
	}
	sessionInfo.Id = id

	var startTime int64
	err = binary.Read(reader, binary.LittleEndian, &startTime)
	if err != nil {
		return nil, err
	}
	sessionInfo.Start = time.Unix(0, startTime*int64(time.Millisecond))

	var dumpTime int64
	err = binary.Read(reader, binary.LittleEndian, &startTime)
	if err != nil {
		return nil, err
	}
	sessionInfo.Dump = time.Unix(0, dumpTime*int64(time.Millisecond))

	return sessionInfo, nil
}

func readString(reader *bufio.Reader) (string, error) {
	var bytesNumber uint16
	err := binary.Read(reader, binary.LittleEndian, &bytesNumber)
	if err != nil {
		return "", err
	}

	buffer := make([]byte, bytesNumber)
	_, err = reader.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer[:]), nil
}
