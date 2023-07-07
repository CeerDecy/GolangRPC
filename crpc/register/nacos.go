package register

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// CreateNacosClient 创建nacos客户端
func CreateNacosClient() (naming_client.INamingClient, error) {
	//create clientConfig
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(""), //When namespace is public, fill in the blank string here.
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)
	// At least one ServerConfig
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(
			"101.43.101.59",
			8848,
			constant.WithScheme("http"),
			constant.WithContextPath("/nacos"),
		),
	}
	// Create naming client for service discovery
	cli, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// RegisService 注册服务
func RegisService(client naming_client.INamingClient, serviceName string, host string, port uint64) error {
	_, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        port,
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
	})
	return err
}

// GetInstance 获取服务实例
func GetInstance(client naming_client.INamingClient, serviceName string) (string, uint64, error) {
	instance, err := client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		//GroupName:   "group-a",             // default value is DEFAULT_GROUP
		//Clusters:    []string{"cluster-a"}, // default value is DEFAULT
	})
	if err != nil {
		return "", 0, err
	}
	return instance.Ip, instance.Port, nil
}
