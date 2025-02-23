package conf

import (
	"flag"
	"github.com/Terry-Mao/goim/pkg/ip"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"time"

	"github.com/BurntSushi/toml"
	xtime "github.com/Terry-Mao/goim/pkg/time"
	"github.com/bilibili/discovery/naming"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	// Conf config
	Conf *Config
)

func init() {
	host = ip.InternalIP()
	flag.StringVar(&confPath, "conf", "job-example.toml", "default config path")
}

// Init init config.
func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile(confPath, &Conf)

	serverConfs := make([]constant.ServerConfig, 0)
	for _, s := range Conf.NacosServerConfig {
		sg := constant.NewServerConfig(s.IpAddr, s.Port, constant.WithScheme(s.Scheme), constant.WithContextPath(s.ContextPath))
		serverConfs = append(serverConfs, *sg)
	}
	Conf.ServerConfig = serverConfs
	Conf.ClientConfg = &constant.ClientConfig{
		NamespaceId:         Conf.NacosClientConfig.NamespaceId, //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           Conf.NacosClientConfig.TimeoutMs,
		NotLoadCacheAtStart: Conf.NacosClientConfig.NotLoadCacheAtStart,
		LogDir:              Conf.NacosClientConfig.LogDir,
		CacheDir:            Conf.NacosClientConfig.CacheDir,
		LogLevel:            Conf.NacosClientConfig.LogLevel,
	}
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Env:       &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},
		Comet:     &Comet{RoutineChan: 1024, RoutineSize: 32},
		Room: &Room{
			Batch:  20,
			Signal: xtime.Duration(time.Second),
			Idle:   xtime.Duration(time.Minute * 15),
		},
	}
}

// Config is job config.
type Config struct {
	Env       *Env
	Kafka     *Kafka
	Discovery *naming.Config
	Comet     *Comet
	Room      *Room

	//nacos 注册中心配置
	NacosClientConfig clientConfig
	NacosServerConfig []*serverConfig

	ServerConfig []constant.ServerConfig

	ClientConfg *constant.ClientConfig
}

// Room is room config.
type Room struct {
	Batch  int
	Signal xtime.Duration
	Idle   xtime.Duration
}

// Comet is comet config.
type Comet struct {
	RoutineChan int
	RoutineSize int
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
}
type clientConfig struct {
	NamespaceId         string
	TimeoutMs           uint64
	NotLoadCacheAtStart bool
	LogDir              string
	CacheDir            string
	LogLevel            string
}

type serverConfig struct {
	IpAddr      string
	ContextPath string
	Port        uint64
	Scheme      string
}
