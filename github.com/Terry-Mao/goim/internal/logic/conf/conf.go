package conf

import (
	"flag"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"time"

	xtime "github.com/Terry-Mao/goim/pkg/time"
	"github.com/bilibili/discovery/naming"

	"github.com/BurntSushi/toml"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string
	host      string
	weight    int64

	// Conf config
	Conf *Config
)

func init() {
	flag.StringVar(&confPath, "conf", "comet-example.toml", "default config path")
}

// Init init config.
func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile(confPath, &Conf)
	// nacos 配置
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
		Env: &Env{Region: region, Zone: zone, DeployEnv: deployEnv, Host: host, Weight: weight},
		Discovery: &naming.Config{
			Nodes:  []string{"192.168.1.6:7171"},
			Region: region,
			Zone:   zone,
			Env:    deployEnv,
			Host:   host,
		},

		HTTPServer: &HTTPServer{
			Network:      "tcp",
			Addr:         "3111",
			ReadTimeout:  xtime.Duration(time.Second),
			WriteTimeout: xtime.Duration(time.Second),
		},
		RPCClient: &RPCClient{Dial: xtime.Duration(time.Second), Timeout: xtime.Duration(time.Second)},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              "3119",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		Backoff: &Backoff{MaxDelay: 300, BaseDelay: 3, Factor: 1.8, Jitter: 1.3},
	}
}

// Config config.
type Config struct {
	Env        *Env
	Discovery  *naming.Config
	RPCClient  *RPCClient
	RPCServer  *RPCServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis
	Node       *Node
	Backoff    *Backoff
	Regions    map[string][]string

	//nacos 注册中心配置
	NacosClientConfig clientConfig
	NacosServerConfig []*serverConfig

	ServerConfig []constant.ServerConfig

	ClientConfg *constant.ClientConfig
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
}

// Node node config.
type Node struct {
	DefaultDomain string
	HostDomain    string
	TCPPort       int
	WSPort        int
	WSSPort       int
	HeartbeatMax  int
	Heartbeat     xtime.Duration
	RegionWeight  float64
}

// Backoff backoff.
type Backoff struct {
	MaxDelay  int32
	BaseDelay int32
	Factor    float32
	Jitter    float32
}

// Redis .
type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string
	Addr         string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
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
