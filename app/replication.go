package main

import (
	"fmt"
	"net"
	"strings"
)

var ReplicationInfo map[string]string = map[string]string{
	"role":               "master",
	"master_replid":      "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", // 40 character alphanumeric string
	"master_repl_offset": "0",
}

func ParseReplication(str string, currentPort string) error {
	parts := strings.Split(str, " ")

	if len(parts) < 2 {
		return fmt.Errorf("ERR invalid input '%s' in --replicaof", str)
	}

	host, replicationPort := parts[0], parts[1]

	if host == "" || replicationPort == "" {
		return fmt.Errorf("ERR invalid input '%s' in --replicaof", str)
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(host, replicationPort))
	if err != nil {
		return err
	}

	connection := &Connection{
		conn: conn,
	}

	// handshake (1/3)
	connection.Write("*1\r\n$4\r\nPING\r\n")

	fmt.Println("CLIENT: PING sent")

	data, err := connection.ReadInput()
	if err != nil {
		return err
	}

	fmt.Println("MASTER:", data)

	// handshake (2/3)
	connection.WriteArray("REPLCONF listening-port " + currentPort)

	fmt.Println("CLIENT: REPLCONF listening-port " + currentPort + " sent")

	data, err = connection.ReadInput()
	if err != nil {
		return err
	}

	fmt.Println("MASTER:", data)

	connection.WriteArray("REPLCONF capa psync2")

	fmt.Println("CLIENT: REPLCONF capa psync2")

	data, err = connection.ReadInput()
	if err != nil {
		return err
	}

	fmt.Println("MASTER:", data)

	// handshake (3/3)
	connection.WriteArray("PSYNC ? -1") // PSYNC {replication_id} {offset}

	fmt.Println("CLIENT: PSYNC ? -1")

	data, err = connection.ReadInput()
	if err != nil {
		return err
	}

	fmt.Println("MASTER:", data)

	ReplicationInfo["role"] = "slave"

	return nil
}
