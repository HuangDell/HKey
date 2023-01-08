package server

import (
	"HKey/pkg"
	"bufio"
	"fmt"
	"math/rand"
	"net/rpc"
	"os"
	"time"
)

/**
Node 实现了非分布式环境下的键值系统，可以通过加入集群来转化为分布式的键值系统
*/

const Join = "Server.Join" // join函数

// InfoArgs 节点加入集群的请求参数
type InfoArgs struct {
	Address  string // 节点的地址
	Name     string // 节点name
	Protocol string // 采取的协议
	id       int    // id
}

type Node struct {
	raft   *Raft
	config pkg.Config
	name   string
}

// NewNode 创建新的节点
func NewNode(data_path string, configPath string, me int) *Node {
	var err error
	s := new(Node)
	s.name = "node" + string(rune(me))
	s.config, err = pkg.ParseConfig(configPath)
	s.raft = NewRaft(data_path, s.config.GetAddress(), me)
	rand.Seed(time.Now().Unix())
	fmt.Println("HKey 正在启动...")
	if err != nil {
		panic(err)
	}
	return s
}

func (this *Node) Initialize() {
	fmt.Printf("HKey %s 启动成功，正在监听%s!\n", this.config.Version, this.config.GetAddress())
	fmt.Println("输入控制中心地址以加入集群...")
	reader := bufio.NewReader(os.Stdin)
	args := InfoArgs{
		Name:     this.name,
		Address:  this.config.GetAddress(),
		Protocol: this.config.Protocol,
	} // 加入集群的请求参数
	var reply string

	conn, err := rpc.DialHTTP(this.config.Protocol, this.config.Raft) // 控制中心连接节点 节点上的rpc调用是与键值系统相关的
	if err != nil {
		fmt.Printf("连接集群失败%s\n正在以独立节点形式运行...\n", err)
		this.raft.role = LEADER
	} else {
		err = conn.Call(Join, args, &reply)
		fmt.Println("成功连接集群!")
	}
	for {
		bytes, _, _ := reader.ReadLine()
		input := string(bytes)
		if input == "exit" {
			break
		} else if input == "log" {
			this.raft.PrintLog()
		}
	}
	fmt.Println("bye")
}
