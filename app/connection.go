package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Connection struct {
	conn net.Conn
}

func (c *Connection) Close() {
	c.conn.Close()
}

func (c *Connection) ReadInput() (string, error) {
	reader := bufio.NewReader(c.conn)

	buf := make([]byte, 1024)

	_, err := reader.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (c *Connection) ParseInput() (string, []string, error) {
	str, err := c.ReadInput()
	if err != nil {
		return "", nil, err
	}

	var args []string

	str = strings.TrimSpace(str)

	parts := strings.Split(str, "\r\n")

	if len(parts) < 1 {
		return "", nil, errors.New("invalid input: empty input")
	}

	totalArgs, err := strconv.Atoi(strings.TrimPrefix(parts[0], "*"))
	if err != nil {
		return "", nil, err
	}

	for i := 1; i <= totalArgs*2; i += 2 {
		_, key := parts[i], parts[i+1]
		args = append(args, key)
	}

	command := strings.ToLower(strings.TrimSpace(args[0]))

	if command == "" {
		return "", nil, errors.New("invalid command: empty command")
	}

	if len(args) == 1 {
		return command, nil, nil
	}

	args = args[1:]

	return command, args, nil
}

func (c *Connection) Write(data string) {
	c.conn.Write([]byte(data))
}

func (c *Connection) WriteSimpleString(str string) {
	simpleStr := fmt.Sprintf("+%s\r\n", str)
	c.Write(simpleStr)
}

func (c *Connection) WriteBulkString(data string) {
	resStr := fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)
	c.Write(resStr)
}

func (c *Connection) WriteArray(data string) {
	parts := strings.Split(strings.TrimSpace(data), " ")

	str := fmt.Sprintf("*%d\r\n", len(parts))

	for i := 0; i < len(parts); i++ {
		str += fmt.Sprintf("$%d\r\n%s\r\n", len(parts[i]), parts[i])
	}

	c.Write(str)
}

func (c *Connection) WriteRDBFile(data string) {
	str := fmt.Sprintf("$%d\r\n%s", len(data), data)
	c.Write(str)
}

func (c *Connection) WriteSimpleError(err string) {
	errStr := fmt.Sprintf("-ERR %s\r\n", err)
	c.Write(errStr)
}

func (c *Connection) WriteNull() {
	c.Write("$-1\r\n")
}
