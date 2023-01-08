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
	n := server.NewNode("build/server/n1/data/", "build/server/n1/config.json", 0)
	n.Initialize()
}
