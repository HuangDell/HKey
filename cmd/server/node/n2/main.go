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
	n := server.NewNode("build/server/n2/data/", "build/server/n2/config.json", 1)
	n.Initialize()
}
