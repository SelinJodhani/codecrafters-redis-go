package main

import "time"

type data struct {
	Val          string
	ExpiresAfter int
	CreatedAt    time.Time
}

var Storage = make(map[string]*data)
