package labrpc

import (
	"HKey/internal/storage"
	"HKey/pkg"
	"fmt"
	"net/rpc"
)

type Node struct {
	endname interface{}
	ch      chan reqMsg // copy of Network.endCh
	config  pkg.Config  // 配置信息
}

// NewNode 创建新的节点
func NewNode(serverName string, configPath string) *Node {
	var err error
	fmt.Println("HKey 正在启动...")
	s := new(Node)
	s.endname = serverName
	s.config, err = pkg.ParseConfig(configPath)
	if err != nil {
		panic(err)
	}
	err = rpc.Register(storage.NewHKey())
	if err != nil {
		panic(err)
	}
	fmt.Printf("启动成功，%s正在监听%s\n", s.endname, s.config.GetAddress())
	return s
}
