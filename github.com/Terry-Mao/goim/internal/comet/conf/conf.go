package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/Terry-Mao/goim/pkg/ip"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"strings"
	"time"

	xtime "github.com/Terry-Mao/goim/pkg/time"
	"github.com/bilibili/discovery/naming"
)

var (
	confPath  string
	region    string
	zone      string
	deployEnv string

	host  string
	addrs string

	weight  int64
	offline bool

	debug bool

	// Conf config
	Conf *Config
)

// confPath = "target\\comet.toml"
func init() {
	host = ip.InternalIP()
	flag.StringVar(&confPath, "conf", "comet-example.toml", "default config path")
	flag.BoolVar(&debug, "debug", false, "default debug")
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
		NamespaceId:         Conf.NacosClientConfig.NamespaceId,
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
		Debug: debug,
		Env: &Env{
			Region:    region,
			Zone:      zone,
			DeployEnv: deployEnv,
			Host:      host,
			Weight:    weight,
			Addrs:     strings.Split(addrs, ","),
			Offline:   offline,
		},
		Discovery: &naming.Config{Region: region, Zone: zone, Env: deployEnv, Host: host},

		RPCClient: &RPCClient{
			Dial:    xtime.Duration(time.Second),
			Timeout: xtime.Duration(time.Second),
		},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              ip.InternalIP(),
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		TCP: &TCP{
			Bind:         []string{":3101"},
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Websocket: &Websocket{
			Bind: []string{":3102"},
		},
		Protocol: &Protocol{
			Timer:            32,
			TimerSize:        2048,
			CliProto:         5,
			SvrProto:         10,
			HandshakeTimeout: xtime.Duration(time.Second * 5),
		},
		Bucket: &Bucket{
			Size:          32,
			Channel:       1024,
			Room:          1024,
			RoutineAmount: 32,
			RoutineSize:   1024,
		},
	}
}

// Config is comet config.
type Config struct {
	Debug     bool
	Env       *Env
	Discovery *naming.Config
	TCP       *TCP
	Websocket *Websocket
	Protocol  *Protocol
	Bucket    *Bucket
	RPCClient *RPCClient
	RPCServer *RPCServer
	Whitelist *Whitelist

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
	Offline   bool
	Addrs     []string
}

type NacosServerConfig struct {
	IpAddr      string
	ContextPath string
	Port        int64
	Scheme      string
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

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
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
