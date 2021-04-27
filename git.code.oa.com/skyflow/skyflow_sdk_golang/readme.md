## skyflow  golang sdk
## 用法

demo 代码在 https://git.code.oa.com/skyflow/skyflow_sdk_golang/blob/master/demo/demo1.go 

#### 1. 定义一个Activty 

```golang 
func checkinput(input string, token string, sf *skyflow.SkyFlow) error {
    var err error
    fmt.Println(input, token)
    // 发送执行成功的消息
    err = sf.SendTaskSuccess(token, input)
    return err
}
```

一个Activity 是一个golang 函数，他是一个 `type ActivityFunction func(string, string, *SkyFlow) error` 类型的函数， 三个参数分别是 :

1. Activity 实例的Input, 是一个活动的执行输入信息的， 类型string ，可以转码成struct。 
2. Activity 实例的Token， 是一个活动实例的具体ID。 
3. skyflow api 对象， 活动执行中用该对象与skyflow server 进行交互。

#### 2. 创建worker ，并启动
demo代码

```golang
address := "http://9.xxx.xxx.xx:80/"     // skyflow server 地址
client, err := skyflow.NewSkyFlow(address)    // 创建一个skyflow api对象
if err != nil {
    fmt.Println(err.Error())
    return
}
// 创建一个ActivityWorker, 需要参数: 1. skyflow api client  2. worker 中的并发数(go runtine 数量) 3. N个WorkerRepository 列表。 每个N个WorkerRepository 都代表一个skyflow 仓库. 
aw, err := skyflow.NewActivityWorker(client, 1000, 
    // NewWRepository 创建一个仓库，一个仓库中包含状态机描述(StateMachine) 和活动 (Activity)，需要参数:
    //  1. 仓库名 . 
    //  2. statemache 路径列表，扫描这些路径下，查找json扩展名文件，准备注册到工作流中. 
    //  3. N个WorkerActivity . 每个WorkerActivity 都是一个函数注册。 
    skyflow.NewWRepository("testrepo", []string{"demo"},
        //NewWActivity 创建一个WorkerActivity , 需要3个参数: 1. 活动名称， 2. 回调函数名称。 3. 活动文档说明。
		skyflow.NewWActivity("add", addActivity, addActivityDoc),
        skyflow.NewWActivity("checkresult", checkresult, checkresultDoc),
    ),
)
// 注册到工作流中
aw.Register()
// 启动worker， 开始轮训查找activity， 并且调用执行。
aw.Run()
// 等10s钟
time.Sleep(10 * time.Second)
// 停止worker
aw.Stop()
```

#### 3. Activity 输入管理

```golang 
func checkinput(input string, token string, sf *skyflow.SkyFlow) error {
	var err error
	fmt.Println(input, token)

    //定义本活动可以接收的参数列表 .post tag 可以管理参数格式要求
	type InputFormat struct {
		X int `json:"x" post:"required"`  // required : 这个参数必需存在，
		Y int `json:"y" post:"required notzero"`  //  notzero 非 0值 
	}

    // 创建参数对象，并且管理默认参数
	var inputdata = InputFormat{
		X: 3,   // 默认值
    }
    // 解析参数， 通过返回检查Input是否满足参数格式的需求。
	err = skyflow.UnmarshalInput(input, &inputdata)
	if err != nil {
        // 直接return err ， worker 会catch这个错误，并给skyflow 发送 ActivityRunError 
		return err
	}

	sum := inputdata.X + inputdata.Y
    // 发送活动的直接结果给活动实例， 使用实例的token
	err = sf.SendTaskSuccess(token, sum)
	return err
}
```