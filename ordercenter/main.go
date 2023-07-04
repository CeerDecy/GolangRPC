package main

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/ordercenter/router"
)

func main() {
	engine := crpc.DefaultEngine()
	router.Router(engine)
	engine.Run(":8002")
}
