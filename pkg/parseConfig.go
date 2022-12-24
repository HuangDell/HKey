package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Protocol string `json:"protocol"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Version  string `json:"version"`
}

// GetAddress 获得完整地址
func (this *Config) GetAddress() string {
	return this.Ip + ":" + this.Port
}

func ParseConfig(configPath string) (Config, error) {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("找不到配置文件,%s", err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var address Config
	json.Unmarshal(byteValue, &address)
	return address, nil
}
