package router

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/goodscenter2/service/goods"
)

func Router(engine *crpc.Engine) {
	group := engine.CreateGroup("goods")
	group.Get("/find", goods.Find)
}
