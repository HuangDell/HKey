package main

import (
	"HKey/internal/server"
	"fmt"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	n := server.NewNode("n1", "build/server/n1/config.json")
	n.Initialize()
}
