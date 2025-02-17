package testTasks

import (
	"fmt"
	"reflect"
	"strings"
)

// 你可以在这里添加你需要的方法
type task struct{}

func Initialize() {
	// 遍历 acase 包中的所有函数并自动调用以 Acase 开头的函数
	callAcaseFunctions()
}

func callAcaseFunctions() {
	v := reflect.ValueOf(&task{}) // 反射获取 acase 类型的指针
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Method(i)
		if strings.HasPrefix(v.Type().Method(i).Name, "Task") {
			// 打印正在执行的函数名
			fmt.Printf("Executing %s\n", v.Type().Method(i).Name)
			method.Call(nil) // 调用方法
		}
	}
}
