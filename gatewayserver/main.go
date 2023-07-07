package main

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/crpc/gateway"
	"net/http"
)

func main() {
	engine := crpc.DefaultEngine()
	engine.OpenGateway = true
	engine.SetGatewayConfig([]*gateway.GWConfig{
		{
			Name:        "order",
			Path:        "/order/**",
			ServiceName: "order",
			//Host: "127.0.0.1",
			//Port: 8888,
			Header: func(req *http.Request) {
				req.Header.Set("Hello", "ssss")
			},
		}, {
			Name:        "goods",
			Path:        "/goods/**",
			ServiceName: "goods",
			//Host: "127.0.0.1",
			//Port: 8001,
		},
	})
	engine.Run(":8000")
}
