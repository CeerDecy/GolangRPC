package main

import (
	"encoding/gob"
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/crpc/rpc"
	"github/CeerDecy/RpcFrameWork/goodscenter/model"
	"github/CeerDecy/RpcFrameWork/goodscenter/router"
	"github/CeerDecy/RpcFrameWork/goodscenter/service"
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

	//server, err := rpc.NewGrpcServer(":9000")
	//if err != nil {
	//	engine.Logger.Error("Main", err.Error())
	//}
	//server.Register(func(g *grpc.Server) {
	//	api.RegisterGoodsApiServer(g, &service.GoodsService{})
	//})
	//err = server.Run()
	//if err != nil {
	//	engine.Logger.Error("Main", err.Error())
	//}
	server := rpc.NewTcpRpcServer(":9000")
	gob.Register(&model.Response{})
	gob.Register(&model.Goods{})
	server.Register("goods", &service.GoodsRpcService{})
	server.Run()
	engine.Run(":8001")
}
