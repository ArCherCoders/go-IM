package main

import (
	"context"
	"flag"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/Terry-Mao/goim/internal/logic/grpc"
	"github.com/Terry-Mao/goim/internal/logic/http"
	"github.com/Terry-Mao/goim/pkg/ip"
	log "github.com/golang/glog"
)

const (
	ver   = "2.0.0"
	appid = "goim.logic"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	// logic
	srv := logic.New(conf.Conf)

	httpSrv := http.New(conf.Conf.HTTPServer, srv)

	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	cancel := registerRpc(srv)

	cancel = registerHttp(srv)

	// 优雅关机

	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if cancel != nil {
				cancel()
			}
			srv.Close()
			httpSrv.Close()
			rpcSrv.GracefulStop()
			log.Infof("goim-logic [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func registerRpc(srv *logic.Logic) context.CancelFunc {
	_, cancel := context.WithCancel(context.Background())

	addr := ip.InternalIP()
	_, port, _ := net.SplitHostPort(conf.Conf.RPCServer.Addr)

	parseport, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		log.Errorf("ip parse port error, err=%+v", err)
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  conf.Conf.ClientConfg,
		ServerConfigs: conf.Conf.ServerConfig,
	})

	instances := vo.RegisterInstanceParam{
		Ip:          addr,
		Port:        parseport,
		ServiceName: appid,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
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
	return cancel
}

func registerHttp(srv *logic.Logic) context.CancelFunc {
	_, cancel := context.WithCancel(context.Background())

	addr := ip.InternalIP()
	_, port, _ := net.SplitHostPort(conf.Conf.HTTPServer.Addr)

	parseport, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		log.Errorf("ip parse port error, err=%+v", err)
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  conf.Conf.ClientConfg,
		ServerConfigs: conf.Conf.ServerConfig,
	})

	instances := vo.RegisterInstanceParam{
		Ip:          addr,
		Port:        parseport,
		ServiceName: "goim.logic.http",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
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
	return cancel
}
