package gojac

import (
	"bufio"
	"os"
	"fmt"
	"encoding/binary"
	"errors"
	"time"
	"io"
)

const (
	headerMagicNumber    uint16 = 0xC0C0
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
		marker, err := reader.ReadByte()
		if err == io.EOF {
			return data, nil
		} else if err != nil {
			return nil, err
		}

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
		}
	}
}

func Write(path string, data ExecutionData) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return err
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	err = writeHeader(writer, data.Version)
	if err != nil {
		return err
	}

	for _, session := range data.Sessions {
		err = writeSessionInfo(writer, session)

		if err != nil {
			return err
		}
	}

	for _, entry := range data.Entries {
		err = writeExecutionEntry(writer, entry)

		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func writeHeader(writer *bufio.Writer, version int16) error {
	err := writer.WriteByte(headerMarker)
	if err != nil {
		return err
	}

	binary.Write(writer, binary.LittleEndian, headerMagicNumber)
	if err != nil {
		return err
	}

	binary.Write(writer, binary.LittleEndian, version)
	if err != nil {
		return err
	}

	return nil
}

func writeSessionInfo(writer *bufio.Writer, info SessionInfo) error {
	err := writer.WriteByte(sessionInfoMarker)
	if err != nil {
		return err
	}

	err = writeString(writer, info.Id)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, info.Start.UnixNano()/int64(time.Millisecond))
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, info.Dump.UnixNano()/int64(time.Millisecond))
	if err != nil {
		return err
	}

	return nil
}

func writeExecutionEntry(writer *bufio.Writer, entry ExecutionEntry) error {
	err := writer.WriteByte(executionEntryMarker)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, entry.Id)
	if err != nil {
		return err
	}

	err = writeString(writer, entry.Name)
	if err != nil {
		return err
	}

	err = writeBooleanArray(writer, entry.Probes)
	if err != nil {
		return err
	}

	return nil
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

	probes, err := readBooleanArray(reader)
	if err != nil {
		return nil, err
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

func readVarInt(reader *bufio.Reader) (int, error) {
	nextByte, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	value := 0xFF & int(nextByte)
	if (value & 0x80) == 0 {
		return value, nil
	}

	nextPart, err := readVarInt(reader)
	if err != nil {
		return 0, err
	}

	return (value & 0x7F) | (nextPart << 7), nil
}

func readBooleanArray(reader *bufio.Reader) ([]bool, error) {
	probesCount, err := readVarInt(reader)
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

	return probes, nil
}

func readString(reader *bufio.Reader) (string, error) {
	var bytesNumber uint16
	err := binary.Read(reader, binary.BigEndian, &bytesNumber)
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

func writeString(writer *bufio.Writer, string string) (error) {
	bytes := []byte(string)
	bytesNumber := uint16(len(bytes))
	binary.Write(writer, binary.BigEndian, bytesNumber)
	writer.Write(bytes)
	return nil
}

func writeBooleanArray(writer *bufio.Writer, array []bool) (error) {
	arrayLength := len(array)
	err := writeVarInt(writer, arrayLength)
	if err != nil {
		return err
	}

	var buffer byte = 0
	var bufferSize uint = 0
	for _, b := range array {
		if b {
			buffer |= 0x01 << bufferSize
		}
		bufferSize++
		if bufferSize == 8 {
			writer.WriteByte(buffer)
			buffer = 0
			bufferSize = 0
		}
	}
	if bufferSize > 0 {
		writer.WriteByte(buffer)
	}
	return nil
}

func writeVarInt(writer *bufio.Writer, value int) error {
	if (value & 0xFFFFFF80) == 0 {
		err := writer.WriteByte(byte(value))
		if err != nil {
			return err
		}
	} else {
		err := writer.WriteByte(byte(0x80 | (value & 0x7F)))
		if err != nil {
			return err
		}
		value = value >> 7
		writeVarInt(writer, value)
	}
	return nil
}
