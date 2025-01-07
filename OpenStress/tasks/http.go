package tasks

import "fmt"

// Task 任务结构体
type Task struct {
	ID      string
	Execute func() // 任务执行函数
}

// Task_HTTP 任务示例
func (t *Task) Task_HTTP() {
	fmt.Println("HTTP Task executed")
}

// Task_TCP 任务示例
func (t *Task) Task_TCP() {
	fmt.Println("TCP Task executed")
}

// HTTP_CLIENT 示例函数，不作为任务
func HTTP_CLIENT() {
	fmt.Println("TEST executed")
}
