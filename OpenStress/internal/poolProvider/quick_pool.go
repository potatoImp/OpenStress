package poolProvider

import (
	"OpenStress/configs"
	"OpenStress/pool"
)

func InitializePool() *pool.Pool {
	// 获取 BaseDetails
	baseDetails := configs.GetBaseDetails()
	return pool.NewPool(baseDetails.Users)
}
