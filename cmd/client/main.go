package main

import (
	HKey "HKey/internal/client"
	"HKey/pkg"
	"bufio"
	"fmt"
	"os"
)

var client HKey.Client

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Println("HKey Client 正在启动...")
	connect()
	command()
}

func connect() {
	config, err := pkg.ParseConfig("./config.json")
	if err != nil {
		panic(err)
	}
	err = client.Connect(config.Protocol, config.Ip, config.Port)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Client %s 连接服务器成功!\n", config.Version)
}

// 获取用户命令
func command() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("HKey>")
		bytes, _, _ := reader.ReadLine()
		input := string(bytes)
		if input == "exit" {
			break
		}
		ans, err := client.Command(input)
		if err != nil {
			fmt.Printf("%v", err)
			break
		}
		fmt.Println(ans)
	}
	fmt.Println("bye")
}

// 处理set命令的解析
func set(words []string) {
	if len(words) != 3 {
		fmt.Println("error arguments")
		return
	}
	err := client.Set(words[1], words[2])
	if err != nil {
		fmt.Println(err)
	}
}

func get(words []string) {
	if len(words) != 2 {
		fmt.Println("error arguments")
		return
	}
	value, err := client.Get(words[1])
	if err != nil {
		fmt.Println(err)
	}
	if value == "nil" {
		fmt.Println("(nil)")
	} else {
		fmt.Printf("\"%s\"\n", value)
	}
}

func del(words []string) {
	if len(words) != 2 {
		fmt.Println("error arguments")
		return
	}
	value, err := client.Del(words[1])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(value)
}

func exists(words []string) {
	if len(words) != 2 {
		fmt.Println("error arguments")
		return
	}
	value, err := client.Exists(words[1])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(value)
}
