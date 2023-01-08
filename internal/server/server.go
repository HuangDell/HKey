package server

import (
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
	config    pkg.Config    // 配置信息
	nodesInfo []InfoArgs    // 记录集群中的节点信息
	peers     []*rpc.Client // 节点
	count     int           // 当前集群节点数
}

func NewServer() *Server {
	var err error
	s := new(Server)
	s.config, err = pkg.ParseConfig("./config.json")
	fmt.Println("HKey 引导程序正在启动...")
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
		bytes, _, _ := reader.ReadLine()
		input := string(bytes)
		if input == "over" {
			this.start()
			break
		}
	}
	fmt.Println("集群启动成功，引导程序将自动关闭。")
}

func (this *Server) start() {
	for _, client := range this.peers {
		err := client.Call("Raft.Start", this.nodesInfo, nil)
		if err != nil {
			panic(err)
		}
	}
}

// Join 供节点远程调用的方法
func (this *Server) Join(args InfoArgs, ans *string) error {
	conn, err := rpc.DialHTTP(args.Protocol, args.Address) // 控制中心同时连接节点
	this.count++
	if err != nil {
		return err
	}
	this.nodesInfo = append(this.nodesInfo, args)
	this.peers = append(this.peers, conn)
	fmt.Printf("节点:%s 地址:%s 成功加入集群！\n", args.Name, args.Address)
	fmt.Printf("集群中节点总数量：%d\n", this.count)
	*ans = "OK"
	return nil
}
