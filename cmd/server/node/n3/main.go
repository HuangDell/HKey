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
	n := server.NewNode("build/server/n3/data/", "build/server/n3/config.json", 2)
	n.Initialize()
}
