/*
 * @Author: mumangtao@gmail.com
 * @Date: 2020-04-06 18:12:22
 * @Last Modified by: mumangtao@gmail.com
 * @Last Modified time: 2020-09-16 20:33:16
 */

package skyflow

import "time"

//ResponseMessage 返回结构体
type ResponseMessage struct {
	ErrorCode    string      `json:"ErrorCode"`
	ErrorMessage string      `json:"ErrorMessage"`
	Success      bool        `json:"Success"`
	Data         interface{} `json:"Data"`
}

// APIRepository  Response 仓库信息
type APIRepository struct {
	Name        string    `json:"Name"`
	Comment     string    `json:"Comment"`
	GmtModified time.Time `json:"GmtModified"`
	GmtCreated  time.Time `json:"GmtCreated"`
}

// APIActivity 活动信息
type APIActivity struct {
	Name        string    `json:"Name"`
	Comment     string    `json:"Comment"`
	Status      string    `json:"Status"`
	URI         string    `json:"URI"`
	GmtModified time.Time `json:"GmtModified"`
	GmtCreated  time.Time `json:"GmtCreated"`
}

// APIActivityTask 查找到的活动信息
type APIActivityTask struct {
	Input     string `json:"Input"`
	TaskToken string `json:"TaskToken"`
}

// APIStateMachine 状态机信息
type APIStateMachine struct {
	Name        string    `json:"Name"`
	Comment     string    `json:"Comment"`
	Content     string    `json:"Content"`
	Status      string    `json:"Status"`
	URI         string    `json:"URI"`
	GmtModified time.Time `json:"GmtModified"`
	GmtCreated  time.Time `json:"GmtCreated"`
}

//APINewExecution 新创建Execution 信息
type APINewExecution struct {
	UUID      string    `json:"UUID"`
	StartTime time.Time `json:"StartTime"`
}

//APIExecution  状态机实例信息
type APIExecution struct {
	ExecutionID     string    `json:"ExecutionID"`
	Title           string    `json:"Title"`
	Input           string    `json:"Input"`
	Output          string    `json:"Output"`
	StateMachineURI string    `json:"StateMachineURI"`
	StartTime       time.Time `json:"StartTime"`
	FinishTime      time.Time `json:"FinishTime"`
	Status          string    `json:"Status"`
}

// ListExecutionCondition API所需要的参数
type ListExecutionCondition struct {
	PageSzie        int    `json:"pageSize"`
	PageNumber      int    `json:"pageNumber"`
	Status          string `json:"status"`
	StateMachineURI string `json:"stateMachineURI"`
	Title           string `json:"title"`
}
