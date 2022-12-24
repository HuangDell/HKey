package server

import (
	"HKey/internal/server/raft"
	"HKey/internal/storage"
	"HKey/pkg"
	"bufio"
	"fmt"
	"net/http"
	"net/rpc"
	"os"
)

/**
定义服务端在rpc过程交换的数据格式
*/

const Join = "Node.Join"       // join函数
const Connect = "Node.Connect" // connect函数

// InfoArgs 节点加入集群的请求参数
type InfoArgs struct {
	Address  string // 节点的地址
	Name     string // 节点name
	Protocol string // 采取的协议
	id       int    // id
}

type Node struct {
	Conn   *rpc.Client // 供控制中心调用的rpc
	config pkg.Config  // 配置信息
	name   string      // 服务器name
	joined bool        // 判断是否已经加入集群
	raft   *raft.Raft
}

// NewNode 创建新的节点
func NewNode(serverName string, configPath string) *Node {
	var err error
	s := new(Node)
	s.config, err = pkg.ParseConfig(configPath)
	s.name = serverName
	s.joined = false
	fmt.Println("HKey 正在启动...")
	if err != nil {
		panic(err)
	}
	return s
}

func (this *Node) Initialize() {

	err := rpc.Register(storage.NewHKey()) // 将storage部分注册为RPC，可供客户端调用
	if err != nil {
		panic(err)
	}
	rpc.HandleHTTP()

	fmt.Printf("HKey %s 启动成功，正在监听%s!\n", this.config.Version, this.config.GetAddress())
	go http.ListenAndServe(this.config.GetAddress(), nil)
	fmt.Println("输入控制中心地址以加入集群...")
	reader := bufio.NewReader(os.Stdin)
	args := InfoArgs{
		Address:  this.config.GetAddress(),
		Protocol: this.config.Protocol,
		Name:     this.name,
	} // 加入集群的请求参数
	var reply string
	for {
		fmt.Printf("HKey> ")
		bytes, _, _ := reader.ReadLine()
		input := string(bytes)
		if input == "exit" {
			break
		} else if this.joined == false {
			conn, err := rpc.DialHTTP(this.config.Protocol, input) // 控制中心连接节点 节点上的rpc调用是与键值系统相关的
			if err != nil {
				fmt.Printf("连接集群失败%s\n", err)
				continue
			}
			err = conn.Call(Join, args, &reply)
			if err != nil {
				fmt.Printf("连接失败%s\n", err)
				continue
			}
			fmt.Println("成功连接集群!")
			this.joined = true
		}
	}
	fmt.Println("bye")
}

// 控制中心调用并启动集群
func (this *Node) run(peersInfo []InfoArgs, ans *string) error {
	peers := make([]*rpc.Client, len(peersInfo))
	for _, arg := range peersInfo {
		if arg.Address == this.config.GetAddress() {
			continue
		}
		conn, err := rpc.DialHTTP(arg.Protocol, arg.Address)
		if err != nil {
			return err
		}
		peers = append(peers, conn)
	}
	return nil
}

// Connect 控制中心调用 用于连接其他节点
func (this *Node) Connect(info InfoArgs, ans *string) error {
	return nil
}
