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
	n := server.NewNode("./data/", "./config.json", 0)
	n.Initialize()
}
