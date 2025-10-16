package config

import (
	"github.com/bpcoder16/Chestnut/v3/contrib/nacos"
	"github.com/bpcoder16/Chestnut/v3/core/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func ParseNacosConfig(dataId, group string, configPtr *AppConfig) (err error) {
	var content string
	content, err = nacos.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		panic(err)
	}
	if err = utils.ParseContentYaml(content, configPtr); err == nil {
		err = configPtr.Check()
	}
	return
}
