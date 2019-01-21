[![GoDoc](https://godoc.org/github.com/tit/go-mango?status.svg)](https://godoc.org/github.com/tit/go-mango)

# Mango VPBX API Wraper

## Install
```bash
$ go get "github.com/tit/go-mango"
```

## Usage
```go
package main

import (
  "crypto/rand"
  "encoding/hex"
  "fmt"
  "go-mango"
  "time"
)

func randomHex(n int) string {
  bytes := make([]byte, n)
  _, _ = rand.Read(bytes)
  return hex.EncodeToString(bytes)
}

func main() {
  client := mango.Client{
    VpbxApiKey:  "",
    VpbxApiSalt: "",
  }

  layout := "2006-01-02 15:04:05"

  fromDateString := "2018-01-20 00:00:00"
  fromDate, _ := time.Parse(layout, fromDateString)

  toDateString := "2018-01-20 23:59:59"
  toDate, _ := time.Parse(layout, toDateString)

  requestId := randomHex(128)
  key, _ := client.StatsKey(fromDate, toDate, requestId)

  time.Sleep(5 * time.Second)

  calls, e := client.Stats(key, requestId)
  fmt.Println(calls, e)
  
  user, e := client.User("666")
  fmt.Println(user, e)
}
```