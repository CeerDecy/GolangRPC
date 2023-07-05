package order

import (
	"context"
	"encoding/json"
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/crpc/rpc"
	"github/CeerDecy/RpcFrameWork/ordercenter/api"
	"github/CeerDecy/RpcFrameWork/ordercenter/model"
	"github/CeerDecy/RpcFrameWork/ordercenter/service"
	"net/http"
)

const tag = "OrderService"

func Find(ctx *crpc.Context) {
	// 以http的方式进行调用
	client := rpc.NewHttpClient()
	client.RegisterHttpService("goods", &service.GoodsService{})
	//params := make(map[string]any)
	//buf, err := client.Get("http://127.0.0.1:8001/goods/find", nil)
	buf, err := client.Do("goods", "Find").(*service.GoodsService).Find(nil)
	if err != nil {
		ctx.Logger.Error(tag, err.Error())
		ctx.JSON(http.StatusOK, model.SuccessResponse(err.Error()))
		return
	}
	var res = new(model.Response)
	err = json.Unmarshal(buf, res)
	ctx.Logger.Info(tag, string(buf))
	ctx.JSON(http.StatusOK, model.SuccessResponse(res.Data))
}

func FindGrpc(ctx *crpc.Context) {
	client, err := rpc.NewGrpcClient(rpc.DefaultGrpcClientConfig("127.0.0.1:9000"))
	if err != nil {
		ctx.Logger.Error("FindGrpc", err.Error())
	}
	defer client.Conn.Close()
	client.Conn.Connect()
	goodsApiClient := api.NewGoodsApiClient(client.Conn)
	goodsResponse, _ := goodsApiClient.Find(context.Background(), &api.GoodsRequest{})
	ctx.JSON(http.StatusOK, goodsResponse)
}
