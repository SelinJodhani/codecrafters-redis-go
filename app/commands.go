package main

var CommandsArgCount = map[string]int{
	"ping":     0, // PING command doesn't requires any arguments
	"echo":     1, // ECHO command requires at least 1 argument
	"get":      1, // GET command requires at least 1 argument
	"set":      2, // SET command requires at least 2 argument
	"info":     1, // INFO command requires at least 1 argument
	"replconf": 2, // REPLCONF command requires at least 2 argument
	"psync":    2, // PSYNC command requires at least 2 argument
}
