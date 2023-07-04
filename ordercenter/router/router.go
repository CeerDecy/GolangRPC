package router

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/ordercenter/service/order"
)

func Router(engine *crpc.Engine) {
	group := engine.CreateGroup("order")
	group.Get("/find", order.Find)
}
