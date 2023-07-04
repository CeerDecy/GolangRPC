package main

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/goodscenter/router"
)

func main() {
	engine := crpc.DefaultEngine()
	router.Router(engine)
	engine.Run(":8001")
}
