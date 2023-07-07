package main

import (
	"encoding/gob"
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/ordercenter/model"
	"github/CeerDecy/RpcFrameWork/ordercenter/router"
)

func main() {
	engine := crpc.DefaultEngine()
	gob.Register(&model.Response{})
	gob.Register(&model.Goods{})
	engine.UseMiddleWare(crpc.Limiter(1, 1))
	//server := rpc.NewTcpRpcServer("127.0.0.1", 8999)
	//server.Register("order", &order.RpcServiceOrder{})
	//server.Run()
	router.Router(engine)
	engine.Run(":8888")
}
