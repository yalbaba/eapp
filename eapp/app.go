package eapp

import (
	"eapp/component"
	"eapp/consts"
	"eapp/logger"
	"eapp/logger/zap"
	"eapp/servers"
	_ "eapp/servers/rpc"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

type IApp interface {
	Start() error
	Stop()
	Rpc(service string, h interface{}) error
}

type App struct {
	c       component.IComponent
	conf    *AppConfigs
	servers map[consts.ServerType]servers.IServer //注册的服务类型
	sync.Mutex
	done bool
}

func NewApp(opts ...Option) IApp {
	conf := &AppConfigs{
		ServerTypes: make(map[consts.ServerType]bool),
	}
	for _, o := range opts {
		o(conf)
	}

	if conf.Log == nil {
		conf.Log = zap.NewDefaultLogger(&logger.LoggerConfig{
			OutputDir: "./logs/",
			OutFile:   "default.log",
			ErrFile:   "default.err",
		})
	}

	app := &App{
		servers: make(map[consts.ServerType]servers.IServer),
		conf:    conf,
		c:       component.NewComponent(conf.Log),
	}

	//初始化所有服务器
	for k, v := range app.conf.ServerTypes {
		app.Lock()
		if v {
			app.servers[k] = servers.NewServer(k, app.c)
		}
		app.Unlock()
	}

	return app
}

func (a *App) Rpc(service string, h interface{}) error {
	if a.servers[consts.RpcServer] == nil {
		return fmt.Errorf("Rpc服务器未初始化")
	}

	return a.servers[consts.RpcServer].RegisterService(service, h)
}

func (a *App) Start() error {

	a.c.Warn("服务器启动中...")

	//开启所有服务器
	for _, server := range a.servers {
		err := server.Start()
		if err != nil {
			return err
		}
	}

	//监听信号
	signalCh := make(chan os.Signal, 1)
	closeCh := make(chan bool)
	signal.Notify(signalCh, os.Interrupt, os.Kill)
	go func() {
		for _ = range signalCh {
			a.c.Warn("正在关闭服务器...")
			a.Stop()
			closeCh <- true
		}
	}()
	<-closeCh //阻塞进程

	return nil
}

func (a *App) Stop() {
	a.done = true
	for _, server := range a.servers {
		server.Stop()
	}
}
