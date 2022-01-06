package engine

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"net/http"
)

type NacosConfig struct {
	NacosServerIP   string //"192.168.6.130"
	NacosServerPort uint64 //8848
	LocalIP         string //"192.168.2.190"
	LocalPort       uint64 //33065
	NacosLogPath    string //"./log"
	ConfigDataID    string //"MTS"
	ConfigGroupName string //"MTS"
	Version         string //"3.0.0.0"

	clientConfig  constant.ClientConfig
	serverConfigs []constant.ServerConfig
	namingClient  naming_client.INamingClient
	configClient  config_client.IConfigClient
}

func (n *NacosConfig) Init() error {
	n.clientConfig = constant.ClientConfig{
		NamespaceId:         "public", // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              n.NacosLogPath + "/nacosLog",
		CacheDir:            n.NacosLogPath + "/nacosCache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	n.serverConfigs = []constant.ServerConfig{
		{
			IpAddr:      n.NacosServerIP,
			ContextPath: "/nacos",
			Port:        n.NacosServerPort,
			Scheme:      "http",
		}, //至少有一个
	}

	var err error
	n.namingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &n.clientConfig,
			ServerConfigs: n.serverConfigs,
		},
	)
	if err != nil {
		return err
	}

	n.configClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &n.clientConfig,
			ServerConfigs: n.serverConfigs,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	n := NacosConfig{
		NacosServerIP:   "192.168.6.130",
		NacosServerPort: 8848,
		LocalIP:         "192.168.2.190",
		LocalPort:       3065,
		NacosLogPath:    "./log",
		ConfigDataID:    "MTS",
		ConfigGroupName: "MTS",
		Version:         "3.0.0.0",
	}
	n.Init()

	conf, _ := n.getConfig()
	fmt.Println(conf)

	inss, _ := n.getInstances()
	for i := 0; i < len(inss); i++ {
		fmt.Println(inss[i])
	}
	n.register()

	h := http.Server{}
	h.ListenAndServe()
	h.Close()
}

// 注册
func (n *NacosConfig) register() (bool, error) {
	return n.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          n.LocalIP,
		Port:        n.LocalPort,
		ServiceName: "MTS", //自己的服务名
		Weight:      1,     //权重
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"version": n.Version}, //额外信息 “version=3.0.0.17”
		ClusterName: "DEFAULT",                               // 默认值DEFAULT
		GroupName:   "DEFAULT",                               // 默认值DEFAULT_GROUP
	})
}

// 获取配置
func (n *NacosConfig) getConfig() (string, error) {
	return n.configClient.GetConfig(vo.ConfigParam{
		DataId: n.ConfigDataID,
		Group:  n.ConfigGroupName})
}

// 获取实例列表
func (n *NacosConfig) getInstances() ([]model.Instance, error) {
	// SelectInstances 只返回满足这些条件的实例列表：healthy=${HealthyOnly},enable=true 和weight>0
	return n.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: "VMS",
		GroupName:   "DEFAULT",           // 默认值DEFAULT_GROUP
		Clusters:    []string{"DEFAULT"}, // 默认值DEFAULT
		HealthyOnly: true,
	})
}

// 获取一个随机健康实例
func (n *NacosConfig) getOneHealthyInstance() (*model.Instance, error) {
	// 实例必须满足的条件：health=true,enable=true and weight>0
	return n.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: "VMS",
		GroupName:   "DEFAULT",           // 默认值DEFAULT_GROUP
		Clusters:    []string{"DEFAULT"}, // 默认值DEFAULT
	})

}
