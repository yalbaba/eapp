# eapp

## RPC服务构建示例
以etcd为注册中心的rpc服务器项目

1、编写自己的handler：

	type testhandler struct {
	}

	type testhandler2 struct {
	}

	func (*testhandler) Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
		fmt.Println("执行rpc任务1")
		//进行链路追踪使用header
		return fmt.Sprintf("input1:::%+v", input), nil
	}

	func (*testhandler2) Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
		fmt.Println("执行rpc任务2")
		//进行链路追踪使用header
		return fmt.Sprintf("input2:::%+v", input), nil
	}

2、配置文件配置（采用toml格式,注意配置文件要放到入口函数同级的configs文件夹下）：

	[registry]
	end_points = ["127.0.0.1:2389"]
	user_name = ""
	password = ""
	register_timeout = 10
	ttl = 8

	[grpc_service]
	cluster = "default"
	port = "9091"
	rpc_timeout = 3
	#10:无权重轮询，20:无权重随机
	balance_mod = 10


3、构建服务器：

	app := eapp.NewApp(
		eapp.WithRpcServer(),
	)

	err := app.RegisterRpcService("yal-test", &testhandler{})
	if err != nil {
		panic(err)
	}
	app.RegisterRpcService("yal-test2", &testhandler2{})
	app.Start()
  
4、首先构建一个连接同一个etcd的app，然后根据服务名和集群名来调用：

	app := eapp.NewApp()
	res, err := app.GetContainer().
		RpcRequest("",
			"yal-test",
			map[string]string{},
			map[string]interface{}{
				"id":   1,
				"name": "yyy",
			}, true)
	if err != nil {
		app.GetContainer().Errorf("err::%v", err)
	}
	res2, err := app.GetContainer().
		RpcRequest("default",
			"yal-test2",
			map[string]string{},
			map[string]interface{}{
				"id":   2,
				"name": "yyy2",
			}, true)
	fmt.Println("res:::::", res)
	fmt.Println("res2:::::", res2)
	
	
## MQC服务器构建示例
1、编写mqc消费服务（框架的实现是nsq，所以这里实现nsq的处理接口）
	type testMqcHandler struct {
		c component.Container
	}

	func NewMqcHandler(c component.Container) *testMqcHandler {
		return &testMqcHandler{c: c}
	}

	func (t *testMqcHandler) HandleMessage(msg *nsq.Message) error {
		t.c.Warn("这是消息:::::", string(msg.Body))
		return nil
	}
	
2、构建app注册消息服务
	app := app.NewApp(
		app.WithMqcServer(),
	)
	app.RegisterMqcService("yangal", "cha1", NewMqcHandler(app.GetContainer()))
	app.Start()
	
3、发送消息
	myapp := app.NewApp(
		app.WithMqcServer(),
	)

	myapp.GetContainer().Send("yangal", map[string]interface{}{
		"name": "yangal",
	})
	
