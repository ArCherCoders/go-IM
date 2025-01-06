package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Terry-Mao/goim/internal/comet/grpc"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Terry-Mao/goim/internal/comet"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	md "github.com/Terry-Mao/goim/internal/logic/model"
	"github.com/Terry-Mao/goim/pkg/ip"
	log "github.com/golang/glog"
)

const (
	ver   = "2.0.0"
	appid = "goim.comet"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Infof("goim-comet [version: %s env: %+v] start", ver, conf.Conf.Env)

	srv := comet.NewServer(conf.Conf)

	if err := comet.InitWhitelist(conf.Conf.Whitelist); err != nil {
		panic(err)
	}
	//InitTCP
	if err := comet.InitTCP(srv, conf.Conf.TCP.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	//InitWebsocket
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	//
	if conf.Conf.Websocket.TLSOpen {
		if err := comet.InitWebsocketWithTLS(srv, conf.Conf.Websocket.TLSBind, conf.Conf.Websocket.CertFile, conf.Conf.Websocket.PrivateFile, runtime.NumCPU()); err != nil {
			panic(err)
		}
	}

	grpc.New(conf.Conf.RPCServer, srv) // 启动grpc 服务

	cancel := register(conf.Conf, srv) // 注册grpc 服务

	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		fmt.Println("阻塞")
		log.Infof("goim-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if cancel != nil {
				cancel()
			}
			srv.Close()
			log.Infof("goim-comet [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

// 服务注册
func register(config *conf.Config, srv *comet.Server) context.CancelFunc {

	_, cancel := context.WithCancel(context.Background())
	env := conf.Conf.Env
	addr := ip.InternalIP()

	_, port, _ := net.SplitHostPort(conf.Conf.RPCServer.Addr)

	parseport, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		log.Errorf("ip parse port error, err=%+v", err)
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  config.ClientConfg,
		ServerConfigs: config.ServerConfig,
	})

	instances := vo.RegisterInstanceParam{
		Ip:          addr,
		Port:        parseport,
		ServiceName: appid,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			md.MetaWeight:  strconv.FormatInt(env.Weight, 10),
			md.MetaOffline: strconv.FormatBool(env.Offline),
			md.MetaAddrs:   strings.Join(env.Addrs, ","),
		},
		ClusterName: "DEFAULT",       // default value is DEFAULT
		GroupName:   "DEFAULT_GROUP", // default value is DEFAULT_GROUP
	}

	success, err := namingClient.RegisterInstance(instances)
	if err != nil {
		panic(err)
		return cancel
	}
	if !success {
		log.Infof("register instance fail, success=%+v", success)
	}

	updateinstance := vo.UpdateInstanceParam{
		Ip:          addr,
		Port:        parseport,
		ServiceName: appid,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			md.MetaWeight:  strconv.FormatInt(env.Weight, 10),
			md.MetaOffline: strconv.FormatBool(env.Offline),
			md.MetaAddrs:   strings.Join(env.Addrs, ","),
		},
		ClusterName: "DEFAULT",       // default value is DEFAULT
		GroupName:   "DEFAULT_GROUP", // default value is DEFAULT_GROUP
	}

	go func() {
		for {
			var (
				err   error
				conns int
				ips   = make(map[string]struct{})
			)
			for _, bucket := range srv.Buckets() {
				for ip := range bucket.IPCount() {
					ips[ip] = struct{}{}
				}
				conns += bucket.ChannelCount()
			}
			updateinstance.Metadata[md.MetaConnCount] = fmt.Sprint(conns)
			updateinstance.Metadata[md.MetaIPCount] = fmt.Sprint(len(ips))

			_, err = namingClient.UpdateInstance(updateinstance)
			if err != nil {
				log.Errorf("dis.Set(%+v) error(%v)", updateinstance, err)
				time.Sleep(time.Second)
				continue
			}
			time.Sleep(time.Second * 10)
		}
	}()
	return cancel
}
