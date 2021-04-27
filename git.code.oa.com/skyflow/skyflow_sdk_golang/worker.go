/*
 * @Author: mumangtao@gmail.com
 * @Date: 2020-04-06 14:16:35
 * @Last Modified by: mumangtao@gmail.com
 * @Last Modified time: 2020-09-16 20:31:32
 */
package skyflow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/panjf2000/ants"
)

// ActivityWorker Worker for activity model
type ActivityWorker struct {
	skyflow      *SkyFlow
	Repositories []*WorkerRepository
	stop         bool
	workerPool   *ants.PoolWithFunc
	poolsize     int
}

// AsyncTaskWorker  worker for skyflow mq
type AsyncTaskWorker struct {
	skyflow      *SkyFlow
	Repositories []*WorkerRepository
	stop         bool
}

// ActivityTaskRuntime    一个具体的Activity 运营时实例
type ActivityTaskRuntime struct {
	activity *WorkerActivity
	data     *APIActivityTask
}

var ActivityRunError = "ActivityRunError"

// NewActivityWorker new activity woker
func NewActivityWorker(client *SkyFlow, poolsize int, repos ...*WorkerRepository) (*ActivityWorker, error) {
	var err error
	aw := ActivityWorker{
		skyflow:      client,
		Repositories: repos,
		stop:         false,
	}
	workerPool, err := ants.NewPoolWithFunc(poolsize, aw.runActivity)
	if err != nil {
		return nil, err
	}
	aw.workerPool = workerPool
	return &aw, nil
}

func (w *ActivityWorker) monitorActivity(act *WorkerActivity) {

	for {
		if w.stop {
			break
		}
		acttask, err := w.skyflow.GetActivityTask(act.URI)
		if err != nil {
			sferr, ok := err.(SkyFlowError)
			if !ok {
				continue
			}
			if sferr.ErrorCode == SkyFlowErrorCode.ActivityTaskNotFound {
				time.Sleep(1 * time.Second)
				continue
			}
		}

		data := ActivityTaskRuntime{
			activity: act,
			data:     &acttask,
		}
		err = w.workerPool.Invoke(data)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

	}

}

func (w *ActivityWorker) runActivity(i interface{}) {

	var err error
	actrun, ok := i.(ActivityTaskRuntime)
	if !ok {
		return
	}
	var token = actrun.data.TaskToken
	defer func() {

		var r = recover()
		if r != nil {
			message := fmt.Sprint(r)
			err = w.skyflow.SendTaskFailure(token, ActivityRunError, message)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}()
	err = actrun.activity.Function(actrun.data.Input, actrun.data.TaskToken, w.skyflow)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

}

// Run  Start Monitor skyflow and execute function when need
func (w *ActivityWorker) Run() error {
	var err error
	err = w.Register()
	if err != nil {
		return err
	}

	for _, repo := range w.Repositories {
		for _, act := range repo.Activitys {
			go w.monitorActivity(act)
		}
	}
	return nil
}

// Register register statemachine / activitys to skyflow server
func (w *ActivityWorker) Register() error {

	var err error
	v, err := w.skyflow.GetAPIVersion()
	if err != nil {
		return err
	}
	fmt.Printf("Register Worker,  Skyflow API Version : %s\n", v)

	for _, repo := range w.Repositories {
		err = repo.ScanStateMachinePath()
		if err != nil {
			return err
		}
		fmt.Println(repo.StateMachines)
		err = w.skyflow.CreateRepository(repo.Name)
		if err != nil {
			return err
		}
		for _, act := range repo.Activitys {
			apiact, err := w.skyflow.CreateActivity(repo.Name, act.Name, act.Comment)
			if err != nil {
				return err
			}
			act.URI = apiact.URI
		}
		for name, content := range repo.StateMachines {
			_, err = w.skyflow.CreateStateMachine(repo.Name, name, content, "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Stop Stop Monitor skyflow
func (w *ActivityWorker) Stop() error {
	w.stop = true
	for {
		if w.workerPool.Running() == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

// WorkerRepository  repository in worker
type WorkerRepository struct {
	Name              string
	Activitys         []*WorkerActivity
	StateMachinePaths []string
	StateMachines     map[string]string
}

// NewWRepository New worker Repository
func NewWRepository(name string, statmachinepaths []string, acts ...*WorkerActivity) *WorkerRepository {

	wr := WorkerRepository{
		Name:              name,
		Activitys:         acts,
		StateMachinePaths: statmachinepaths,
		StateMachines:     map[string]string{},
	}
	return &wr
}

// ScanStateMachinePath  add statemachine search path
func (wr *WorkerRepository) ScanStateMachinePath() error {

	for _, p := range wr.StateMachinePaths {
		finfo, err := os.Stat(p)
		if err != nil {
			return err
		}
		if finfo.IsDir() {
			// 扫描目录
			fileinfos, err := ioutil.ReadDir(p)
			if err != nil {
				return err
			}
			// 遍历文件
			for _, fi := range fileinfos {
				fmt.Println(fi.Name())
				if !fi.IsDir() {
					finame := fi.Name()
					fnamepart := strings.SplitN(strings.TrimSpace(finame), ".", 2)
					if len(fnamepart) < 2 {
						continue
					}
					fname := fnamepart[0]
					fext := fnamepart[1]
					if fext == "json" {
						fullfilename := path.Join(p, finame)
						content, err := ioutil.ReadFile(fullfilename)
						if err != nil {
							return err
						}
						wr.StateMachines[fname] = string(content)
					}
				}
			}
		} else {
			// 处理文件
			finame := finfo.Name()
			fnamepart := strings.SplitAfterN(strings.TrimSpace(finame), ".", 2)
			fname := fnamepart[0]
			fext := fnamepart[1]
			if fext == "json" {
				content, err := ioutil.ReadFile(finame)
				if err != nil {
					return err
				}
				wr.StateMachines[fname] = string(content)
			}

		}
	}
	return nil
}

// WorkerActivity Activity in worker
type WorkerActivity struct {
	Name     string           //  活动名称
	Function ActivityFunction // 指定的函数
	Comment  string           // 活动说明
	URI      string           // 活动URI
}

// ActivityFunction activity callback function
type ActivityFunction func(string, string, *SkyFlow) error

// NewWActivity New Worker Activity
func NewWActivity(name string, f ActivityFunction, comment string) *WorkerActivity {
	wa := WorkerActivity{
		Name:     name,
		Function: f,
		Comment:  comment,
	}
	return &wa
}

// UnmarshalInput 解压传来的到特定的变量钟
// @input  string ,要处理的输入的值
// @v   interface, 要解压的目标数据结构，
// v struct tag :   `json:"x", post:"notzero, required" `
//      notzero : 字段值非0值
//      required: 字段值必需要有显式声明
func UnmarshalInput(input string, v interface{}) error {
	var err error
	content := []byte(input)
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	var mapjson = map[string]interface{}{}
	err = json.Unmarshal(content, &mapjson)
	if err != nil {
		return err
	}

	var isBlank = func(value reflect.Value) bool {
		switch value.Kind() {
		case reflect.String:
			return value.Len() == 0
		case reflect.Bool:
			return !value.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return value.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return value.Uint() == 0
		case reflect.Float32, reflect.Float64:
			return value.Float() == 0
		case reflect.Interface, reflect.Ptr:
			return value.IsNil()
		}
		return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
	}

	t := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("post")

		var key = t.Field(i).Tag.Get("json")
		if key == "" {
			key = t.Field(i).Name
		}
		rv := val.Field(i)

		tagitems := strings.Fields(tag)
		for _, titem := range tagitems {
			switch titem {
			case "required":
				if _, ok := mapjson[key]; !ok {
					message := MessageFormat.ArgumentRequried.Format(key)
					err = fmt.Errorf(message)
					return err
				}
			case "notzero":
				if isBlank(rv) {
					message := MessageFormat.ArgumentZero.Format(key)
					err = fmt.Errorf(message)
					return err
				}
			}
		}
	}
	return nil
}

// MessageTemplate 异常类型信息模板
type MessageTemplate struct {
	Template string
}

// Format  格式话异常信息
func (mt MessageTemplate) Format(params ...interface{}) string {
	msg := fmt.Sprintf(mt.Template, params)
	return msg
}

// MessageFormat  消息类型与异常信息格式化映射
var MessageFormat = struct {
	ArgumentVauleError MessageTemplate
	ArgumentTypeError  MessageTemplate
	ArgumentRequried   MessageTemplate
	ArgumentZero       MessageTemplate
}{
	ArgumentVauleError: MessageTemplate{
		Template: "Argument '%s' Should Not Be '%s'",
	},
	ArgumentTypeError: MessageTemplate{
		Template: "Argument '%s' Should  Be type '%s'",
	},
	ArgumentRequried: MessageTemplate{
		Template: "Argument '%s' Required",
	},
	ArgumentZero: MessageTemplate{
		Template: "Argument '%s' Should Not Be Zero Value",
	},
}
