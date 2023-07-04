package order

import (
	"encoding/json"
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/crpc/rpc"
	"github/CeerDecy/RpcFrameWork/ordercenter/model"
	"net/http"
)

const tag = "OrderService"

func Find(ctx *crpc.Context) {
	// 以http的方式进行调用
	client := rpc.NewHttpClient()
	buf, err := client.Get("http://127.0.0.1:8001/goods/find")
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
