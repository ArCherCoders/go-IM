package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Terry-Mao/goim/internal/job"
	"github.com/Terry-Mao/goim/internal/job/conf"
	log "github.com/golang/glog"
)

var (
	ver = "2.0.0"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Infof("goim-job [version: %s env: %+v] start", ver, conf.Conf.Env)

	j := job.New(conf.Conf)
	go j.Consume()

	// 下面的代码  是进行优雅的关闭  固定写法
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			j.Close()
			log.Infof("goim-job [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
