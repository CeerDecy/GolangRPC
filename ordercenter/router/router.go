package router

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/ordercenter/service/order"
)

func Router(engine *crpc.Engine) {
	group := engine.CreateGroup("order")
	group.Get("/find", order.Find)
	group.Get("/findGrpc", order.FindGrpc)
	group.Get("/findTcp", order.FindTcp)
	//group.Get("/findRpc", order.RpcServiceOrder{}.FindRpc)
}
