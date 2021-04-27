/*
 * @Author: mumangtao@gmail.com
 * @Date: 2020-09-16 17:47:45
 * @Last Modified by: mumangtao@gmail.com
 * @Last Modified time: 2020-09-16 17:48:28
 */

package main

import (
	"fmt"
	"skyflow"
	"time"
)

func main() {

	var err error
	address := "http://9.134.6.17:80/"
	client, err := skyflow.NewSkyFlow(address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	aw, err := skyflow.NewActivityWorker(client, 1000,
		skyflow.NewWRepository("testrepo", []string{"demo"},
			skyflow.NewWActivity("add", addActivity, addActivityDoc),
			skyflow.NewWActivity("checkresult", checkresult, checkresultDoc),
		),
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = aw.Register()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	smuri := "statemachine:testrepo/testact"
	input := map[string]int{
		"x": 1,
		"y": 3,
	}
	apinewexe, err := client.StartExecution(smuri, "test go worker", input)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(apinewexe)
	// 启动worker
	aw.Run()
	// 等10s钟
	time.Sleep(10 * time.Second)
	// 停止worker
	aw.Stop()
}

var addActivityDoc = `
add 计算两个数的和
输入:
	{
		"x" : 1, 
		"y" : 2
	}
输出:

3

`

func addActivity(input string, token string, sf *skyflow.SkyFlow) error {
	var err error
	fmt.Println(input, token)

	type InputFormat struct {
		X int `json:"x" post:"required"`
		Y int `json:"y" post:"required notzero"`
	}

	var inputdata = InputFormat{
		X: 0,
	}
	err = skyflow.UnmarshalInput(input, &inputdata)
	if err != nil {
		return err
	}

	sum := inputdata.X + inputdata.Y

	err = sf.SendTaskSuccess(token, sum)
	return err
}

var checkresultDoc = `dfe 
check input 
输入:
输出:
`

func checkresult(input string, token string, sf *skyflow.SkyFlow) error {
	var err error

	fmt.Println(input, token)

	err = sf.SendTaskSuccess(token, input)
	return err
}
