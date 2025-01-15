package utils

import (
	"OpenStress/pool"
)

// logAndReraise 函数
func logAndReraise(retryState string) {
	logger, err := pool.GetLogger()
	if err != nil {
		logger.Log("ERROR", "Error get logger:"+err.Error())
		return
	}

	// 记录错误日志
	logger.Log("ERROR", "Retry attempts exhausted. Last exception retryState: "+retryState)

	// 记录警告日志
	logger.Log("ERROR", `
Recommend going to https://deepwisdom.feishu.cn/wiki/MsGnwQBjiif9c3koSJNcYaoSnu4#part-XdatdVlhEojeAfxaaEZcMV3ZniQ 
See FAQ 5.8
`)
}
