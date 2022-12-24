package client

import (
	"HKey/internal/storage"
	"fmt"
	"net/rpc"
)

type Client struct {
	conn *rpc.Client
}

// Connect 通过RPC连接服务器
func (this *Client) Connect(protocol string, ip string, port string) error {
	address := ip + ":" + port
	conn, err := rpc.DialHTTP(protocol, address)
	this.conn = conn
	if err != nil {
		return fmt.Errorf("连接服务器失败")
	}
	return nil
}

func (this *Client) Close() {
	this.conn.Close()
}

// Set 插入键或者更新键值
func (this *Client) Set(key string, value string) error {
	var ans string
	var args = storage.SetArgs{
		Key: key, Value: value,
	}
	err := this.conn.Call(storage.SetRPC, args, &ans)
	if err != nil {
		return fmt.Errorf("服务器错误 %s", err)
	}
	return nil
}

func (this *Client) Get(key string) (string, error) {
	var ans string
	err := this.conn.Call(storage.GetRPC, key, &ans)
	if err != nil {
		return "", fmt.Errorf("服务器错误 %s", err)
	}
	return ans, nil
}

func (this *Client) Del(key string) (string, error) {
	var ans string
	err := this.conn.Call(storage.DelRPC, key, &ans)
	if err != nil {
		return "", fmt.Errorf("服务器错误 %s", err)
	}
	return ans, nil
}

func (this *Client) Exists(key string) (string, error) {
	var ans string
	err := this.conn.Call(storage.ExistsRPC, key, &ans)
	if err != nil {
		return "", fmt.Errorf("服务器错误 %s", err)
	}
	return ans, nil
}
