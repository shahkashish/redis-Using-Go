package main

import (
	"strconv"
	"sync"
	"time"
)

type valueWithExpiry struct {
	value  string
	expiry time.Time
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

var SETs = make(map[string]valueWithExpiry)
var SETsMu = sync.RWMutex{}

func set(args []Value) Value {
	var expiry time.Time
	if len(args) == 4 {
		expiryMilliseconds, err := strconv.ParseInt(args[3].bulk, 10, 64)
		if err != nil {
			return Value{typ: "error", str: "Invalid expiry time"}
		}
		expiry = time.Now().Add(time.Duration(expiryMilliseconds) * time.Millisecond)

	} else if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].bulk
	value := valueWithExpiry{
		value:  args[1].bulk,
		expiry: expiry,
	}
	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}
func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value.value}
}
func removeExpiredKeys() {
	for {
		 // Adjust the frequency of expiry checks as needed
		SETsMu.Lock()
		for key, item := range SETs {
			if item.expiry != (time.Time{}) && item.expiry.Before(time.Now()) {
				delete(SETs, key)
			}
		}
		SETsMu.Unlock()
		
		time.Sleep(time.Millisecond)
	}
}

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": ping,
	"SET":  set,
	"GET":  get,
}
