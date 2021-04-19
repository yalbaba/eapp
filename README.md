# eapp
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

	err := app.Rpc("yal-test", &testhandler{})
	if err != nil {
		panic(err)
	}
	app.Rpc("yal-test2", &testhandler2{})
	app.Start()
  
4、根据服务名和集群名来调用：

	app := eapp.NewApp(
		eapp.WithRpcServer(),
	)

	app.Request("default", "yal-test", map[string]string{}, map[string]interface{}{
		"id": 1,
	}, true)
