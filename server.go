package main

import (
	"fmt"
	"io"
	"strings"

	// Uncomment this block to pass the first stage
	//"log"
	"net"
	"os"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}
func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

// *2\r\n$4\r\necho\r\n$3\r\nhey\r\n
//
//	func echoFunc(s string) string {
//		i := 1
//		wordsOfString := 0
//		for s[i] >= '0' && s[i] <= '9' {
//			wordsOfString += wordsOfString*10 + int(s[i]) - '0'
//			i++
//			}
//			echoString := ""
//			wordsOfString -= 1
//			i += 9
//			for wordsOfString > 0 {
//				i += 4
//				len := 0
//				for s[i] >= '0' && s[i] <= '9' {
//					len += len*10 + int(s[i]) - '0'
//					i++
//				}
//				wordsOfString--
//				i += 2
//				echoString += string(s[i : i+len])
//				if wordsOfString > 0 {
//					echoString += " "
//				}
//			}
//			return echoString
//		}
//
// var requestPing = []byte("*1\r\n$4\r\nping\r\n")
// var responsePing = []byte("+PONG\r\n")
func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		if value.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}
		if len(value.array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]
		writer := NewWriter(conn)
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}
		result := handler(args)
		writer.Write(result)
		// switch value.typ {
		// case ARRAY:
		// 	log.Print("Array received:", value.Array)
		// 	if strings.ToLower(value.Array[0].String) == "ping" {
		// 		conn.Write(responsePing)
		// 	} else if strings.ToLower(value.Array[0].String) == "quit" {
		// 		return
		// 	} else if strings.ToLower(value.Array[0].String) == "echo" {
		// 		if len(value.Array) == 2 {
		// 			res := encodeBulkString(value.Array[1].String)
		// 			conn.Write([]byte(res))
		// 		} else {
		// 			conn.Write([]byte("-ERR wrong number of arguments for 'echo' command\r\n"))
		// 		}
		// 	} else if strings.ToLower(value.Array[0].String) == "set" {
		// 		m[value.Array[1].String] = value.Array[2].String
		// 		conn.Write([]byte("+OK\r\n"))
		// 	} else if strings.ToLower(value.Array[0].String) == "get" {
		// 		res, ok := m[value.Array[1].String]
		// 		if ok {
		// 			res = encodeBulkString(res)
		// 			conn.Write([]byte(res))
		// 		} else {
		// 			res = encodeBulkString("")
		// 			conn.Write([]byte(res))
		// 		}
		// 	} else {
		// 		conn.Write([]byte("+OK\r\n"))
		// 	}
		// default:
		// 	conn.Write([]byte("+OK\r\n"))
		// }
	}
}
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	
	go removeExpiredKeys()
	for {
		ln, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(ln)
	}
}
