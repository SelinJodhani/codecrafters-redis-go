package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type RedisServer struct {
	listener net.Listener
}

func NewRedisServer(port string) (*RedisServer, error) {
	l, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	return &RedisServer{
		listener: l,
	}, nil
}

func (rs *RedisServer) Close() {
	rs.listener.Close()
}

func (rs *RedisServer) Serve() {
	defer rs.Close()

	fmt.Println("Server listening on port", rs.listener.Addr().(*net.TCPAddr).Port)

	for {
		conn, err := rs.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		connection := &Connection{
			conn: conn,
		}

		go rs.HandleConnection(connection)
	}
}

func (rs *RedisServer) HandleConnection(conn *Connection) {
	defer conn.Close()

	// infinite loop to keep reading input from same connection
	for {
		command, args, err := conn.ParseInput()

		if err != nil {
			conn.WriteSimpleError(fmt.Sprint("Error parsing input: ", err))
			return
		}

		if minArgs, ok := CommandsArgCount[command]; ok && len(args) < minArgs {
			conn.WriteSimpleError(
				fmt.Sprintf("wrong number of arguments for '%s' command", command),
			)
			continue
		}

		switch command {

		case "ping":
			conn.WriteSimpleString("PONG")

		case "echo":
			str := strings.Join(args, " ")
			conn.WriteBulkString(str)

		case "set":
			key, val := args[0], args[1]

			var pxVal int

			if len(args) > 3 {
				px := strings.ToLower(args[2])
				if px != "px" {
					conn.WriteSimpleError(fmt.Sprintf("unknown command '%s'\r\n", px))
					continue
				}

				pxVal, err = strconv.Atoi(args[3])
				if err != nil {
					conn.WriteSimpleError(
						fmt.Sprintf("unknown value after PX '%s'\r\n", args[3]),
					)
					continue
				}
			}

			Storage[key] = &data{
				Val:          val,
				ExpiresAfter: pxVal,
				CreatedAt:    time.Now(),
			}

			conn.WriteSimpleString("OK")

		case "get":
			key := strings.TrimSpace(args[0])
			data, ok := Storage[key]

			if !ok {
				conn.WriteNull()
				continue
			}

			timePassed := time.Since(data.CreatedAt)

			if data.ExpiresAfter != 0 && timePassed.Milliseconds() > int64(data.ExpiresAfter) {
				conn.WriteNull()
			} else {
				conn.WriteBulkString(data.Val)
			}

		case "info":
			subject := strings.ToLower(args[0])

			info, ok := Info[subject]

			if !ok {
				conn.WriteSimpleError(fmt.Sprintf("unknown arg '%s'", subject))
				continue
			}

			str := fmt.Sprintf("# %s\r\n", subject)

			for key, val := range info {
				str += fmt.Sprintf("%s:%s\r\n", key, val)
			}

			conn.WriteBulkString(str)

		case "replconf":
			conn.WriteSimpleString("OK")

		case "psync":
			str := fmt.Sprintf(
				"FULLRESYNC %s %s",
				ReplicationInfo["master_replid"],
				ReplicationInfo["master_repl_offset"],
			)
			conn.WriteSimpleString(str)

			emptyRdb, err := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
			if err != nil {
				conn.WriteSimpleError(fmt.Sprint("Error creating empty rdb file: ", err))
				return
			}

			conn.WriteRDBFile(string(emptyRdb))

		default:
			conn.WriteSimpleError(fmt.Sprintf("unknown command '%s'", command))
		}
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	port := flag.String("port", "6379", "Server port")
	replicaof := flag.String(
		"replicaof",
		"",
		"Simulate a replica showing commands received from the master.",
	)

	flag.Parse()

	if *replicaof != "" {
		err := ParseReplication(*replicaof, *port)
		if err != nil {
			fmt.Println("Failed to connect to replica server: ", err)
			os.Exit(1)
		}
	}

	server, err := NewRedisServer(":" + *port)
	if err != nil {
		fmt.Println("Failed to start redis server: ", err)
		os.Exit(1)
	}

	server.Serve()
}
