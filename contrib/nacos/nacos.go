package nacos

import (
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	once    sync.Once
	iClient config_client.IConfigClient
)

func Init(serverConfigs []constant.ServerConfig, clientConfig *constant.ClientConfig) {
	once.Do(func() {
		var err error
		iClient, err = clients.NewConfigClient(vo.NacosClientParam{
			ServerConfigs: serverConfigs,
			ClientConfig:  clientConfig,
		})
		if err != nil {
			panic(err)
		}
	})
}

func GetConfig(param vo.ConfigParam) (string, error) {
	return iClient.GetConfig(param)
}
