# gojac
Go package for working with the JaCoCo execution data file format

## Usage

### Reading exec file

```go
executionData, err := Read("coverage.exec")
```

### Writing exec file

```go
sessionInfo := SessionInfo{
    Id:    "some_session",
    Start: time.Now(),
    Dump:  time.Now(),
}
executionEntry := ExecutionEntry{
    Id:     123,
    Name:   "com.java",
    Probes: []bool{true, false, true},
}
executionData := ExecutionData{
    Version:  0x1007,
    Entries:  []ExecutionEntry{executionEntry},
    Sessions: []SessionInfo{sessionInfo},
}

err = Write("coverage.exec", executionData)
```