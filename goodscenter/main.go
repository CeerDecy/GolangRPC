package main

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/crpc/rpc"
	"github/CeerDecy/RpcFrameWork/goodscenter/api"
	"github/CeerDecy/RpcFrameWork/goodscenter/router"
	"github/CeerDecy/RpcFrameWork/goodscenter/service"
	"google.golang.org/grpc"
)

func main() {
	engine := crpc.DefaultEngine()
	router.Router(engine)

	// Grpc
	//server := grpc.NewServer()
	//listen, _ := net.Listen("tcp", ":9000")
	//api.RegisterGoodsApiServer(server, &rpc.GoodsService{})
	//err := server.Serve(listen)
	//if err != nil {
	//	engine.Logger.Error("Main", err.Error())
	//}
	server, err := rpc.NewGrpcServer(":9000")
	if err != nil {
		engine.Logger.Error("Main", err.Error())
	}
	server.Register(func(g *grpc.Server) {
		api.RegisterGoodsApiServer(g, &service.GoodsService{})
	})
	err = server.Run()
	if err != nil {
		engine.Logger.Error("Main", err.Error())
	}
	engine.Run(":8001")
}
