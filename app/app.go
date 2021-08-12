package app

import (
	"eapp/component"
	"eapp/consts"
	"eapp/global_config"
	"eapp/logger"
	"eapp/logger/zap"
	"eapp/servers"
	_ "eapp/servers/mqc/nsq"
	_ "eapp/servers/rpc/grpc"
	"flag"
	"github.com/BurntSushi/toml"
	"log"
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

var configPath string

type App struct {
	c       component.IComponent
	conf    *AppConfigs
	servers map[consts.ServerType]servers.IServer //注册的服务类型
	sync.Mutex
	done bool
}

func NewApp(opts ...Option) IApp {

	initConfig()

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

//命令行参数解析配置文件
func initConfig() {

	flag.StringVar(&configPath, "c", "", "服务器配置文件路径")
	flag.Parse()

	if configPath == "" {
		// log.Warn("未指定配置文件路径! 将使用 ./configs/config_dev.toml 配置文件加载程序")

		configPath = "./configs/config_dev.toml"
	}

	if _, err := toml.DecodeFile(configPath, global_config.Conf); err != nil {
		log.Panic(err)
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
