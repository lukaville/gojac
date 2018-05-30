package gojac

import "time"

type ExecutionEntry struct {
	Id     int64
	Name   string
	Probes []bool
}

type SessionInfo struct {
	Id    string
	Start time.Time
	Dump  time.Time
}

type ExecutionData struct {
	Version  int16
	Entries  []ExecutionEntry
	Sessions []SessionInfo
}
