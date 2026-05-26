package pool

import "errors"

// GetMQConnFunc 获取MQ连接函数类型
type GetMQConnFunc func() (interface{}, error)

// PutMQConnFunc 归还MQ连接函数类型
type PutMQConnFunc func(conn interface{})

// SubmitTaskFunc 提交任务函数类型
type SubmitTaskFunc func(fn func())

var (
	getMQConn     GetMQConnFunc
	putMQConn     PutMQConnFunc
	submitTask    SubmitTaskFunc
)

// SetMQConnGetters 设置MQ连接获取和归还函数
func SetMQConnGetters(getFunc GetMQConnFunc, putFunc PutMQConnFunc) {
	getMQConn = getFunc
	putMQConn = putFunc
}

// SetSubmitTask 设置任务提交函数
func SetSubmitTask(submitFunc SubmitTaskFunc) {
	submitTask = submitFunc
}

// GetMQConn 获取MQ连接
func GetMQConn() (interface{}, error) {
	if getMQConn != nil {
		return getMQConn()
	}
	return nil, errors.New("MQ连接获取函数未设置")
}

// PutMQConn 归还MQ连接
func PutMQConn(conn interface{}) {
	if putMQConn != nil {
		putMQConn(conn)
	}
}

// SubmitTask 提交任务
func SubmitTask(fn func()) {
	if submitTask != nil {
		submitTask(fn)
	}
}
