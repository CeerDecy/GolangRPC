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
	router.Router(engine)
	engine.Run(":8002")
}
