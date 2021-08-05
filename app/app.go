package app

import (
	"eapp/component"
	"eapp/consts"
	"eapp/logger"
	"eapp/logger/zap"
	"eapp/servers"
	_ "eapp/servers/mqc/nsq"
	_ "eapp/servers/rpc/grpc"
	"os"
	"os/signal"
	"sync"
)

type IApp interface {
	Start() error
	Stop()
	GetContainer() component.Container
	RegisterRpcService(service string, sc interface{}) error
	RegisterMqcService(topic, channel string, sc interface{}) error
}

//---------------以下为实现-----------------
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

	a := &App{
		conf:    conf,
		c:       component.NewComponent(conf.Log),
		servers: make(map[consts.ServerType]servers.IServer),
	}

	for k, v := range a.conf.ServerTypes {
		a.Lock()
		if v {
			a.servers[k] = servers.NewServer(k, a.c)
			switch k {
			case consts.RpcServer:
				a.c.NewRpcInvoker()
			case consts.MqcServer:
				a.c.NewMqcProducer()
			}
		}
		a.Unlock()
	}

	return a
}

func (a *App) Start() error {

	a.c.Debug("服务器启动中...")

	for _, server := range a.servers {
		err := server.Start()
		if err != nil {
			return err
		}
	}

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
	<-closeCh

	return nil
}

func (a *App) Stop() {
	a.done = true
	for _, server := range a.servers {
		server.Stop()
	}
}

func (a *App) GetContainer() component.Container {
	return a.c
}

func (a *App) RegisterRpcService(service string, sc interface{}) error {
	if a.servers[consts.RpcServer] == nil {
		panic("Rpc服务器未初始化")
	}

	return a.servers[consts.RpcServer].RegisterService(service, sc)
}

func (a *App) RegisterMqcService(topic, channel string, sc interface{}) error {
	if topic == "" || channel == "" {
		panic("消息队列名不能为空")
	}
	return a.servers[consts.MqcServer].RegisterService(topic+"/"+channel, sc)
}
