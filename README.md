# eapp
以etcd为注册中心的rpc服务器项目

1、编写自己的handler：

	func MyHandler(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
		fmt.Println("执行rpc任务1")
		//进行链路追踪使用header
		return fmt.Sprintf("input1:::%+v", input), nil
	}

	func MyHandler2(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
		fmt.Println("执行rpc任务2")
		//进行链路追踪使用header
		return fmt.Sprintf("input2:::%+v", input), nil
	}

2、配置文件配置（采用toml格式,注意配置文件要放到入口函数同级的configs文件夹下）：

	[registry]
	addrs = ["127.0.0.1:2389"]
	user_name = ""
	password = ""
	register_time_out = 5
	ttl = 8

	[grpc_service]
	cluster = "default"
	port = "9091"
	rpc_time_out = 3


3、构建服务器：

	app, err := server.NewApp(server.WithTimeOut(time.Second),
		server.WithBalancer(11))
	if err != nil {
		fmt.Println("NewErpcServer:::", err)
		return
	}

	app.RegisterService("yal-test", MyHandler)
	app.RegisterService("yal-test2", MyHandler2)
	app.Start()
  
4、根据服务名和集群名来调用：

	app, err := server.NewApp(server.WithTimeOut(time.Second),
		server.WithBalancer(11))
	if err != nil {
		fmt.Println("NewErpcServer:::", err)
		return
	}

	resp, err := app.Rpc("default", "yal-test", map[string]string{}, map[string]interface{}{
		"id":   1,
		"name": "yang",
	}, true)

	fmt.Println("resp:::::::::::", resp)

	resp2, err := app.Rpc("default", "yal-test2", map[string]string{}, map[string]interface{}{
		"id":   2,
		"name": "yang2",
	}, true)

	fmt.Println("resp2::::::::::", resp2)
