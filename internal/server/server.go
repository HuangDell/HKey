package server

import (
	"HKey/internal/server/raft"
	"HKey/pkg"
	"bufio"
	"fmt"
	"net/http"
	"net/rpc"
	"os"
)

/**
服务器集群接入点，当服务器节点都接入时，开始运作。
*/

type Server struct {
	config    pkg.Config      // 配置信息
	nodesInfo []InfoArgs      // 记录集群中的节点信息
	peers     []*rpc.Client   // 节点
	saved     *raft.Persister // 持久化保存
	raft      *raft.Raft
	info      chan raft.ApplyMsg
}

func NewServer() *Server {
	var err error
	s := new(Server)
	s.config, err = pkg.ParseConfig("build/server/config.json")
	fmt.Println("HKey 集群控制中心正在启动...")
	s.info = make(chan raft.ApplyMsg)
	if err != nil {
		panic(err)
	}
	return s
}

func (this *Server) Run() {
	err := rpc.Register(this) // 将storage部分注册为RPC，可供客户端调用
	if err != nil {
		panic(err)
	}
	rpc.HandleHTTP()

	fmt.Printf("HKey %s 启动成功，正在监听%s 并等待服务器加入\n", this.config.Version, this.config.GetAddress())
	go http.ListenAndServe(this.config.GetAddress(), nil)
	fmt.Println("输入over结束准备阶段")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("HKey>")
		bytes, _, _ := reader.ReadLine()
		input := string(bytes)
		if input == "over" {
			this.saved = raft.MakePersister()
			this.raft = raft.NewRaft(this.peers, 0, this.saved, this.info)
		}
	}
}

// 将节点之间互相连接
func (this *Server) nodeConnect() {
	l := len(this.peers)
	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			if i == j {
				continue
			}

		}
	}
}

// Join 供节点远程调用的方法
func (this *Server) Join(args InfoArgs, ans *string) error {
	conn, err := rpc.DialHTTP(args.Protocol, args.Address) // 控制中心同时连接节点
	if err != nil {
		return err
	}
	this.nodesInfo = append(this.nodesInfo, args)
	this.peers = append(this.peers, conn)
	fmt.Printf("HKey> 节点:%s 地址:%s 成功加入集群！\n", args.Name, args.Address)
	*ans = "OK"
	return nil
}

// Command 控制中心处理客户的请求，并将实际操作转移到节点
func (this *Server) Command(command string, ans *string) error {

	return nil
}
