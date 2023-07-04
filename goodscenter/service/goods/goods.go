package goods

import (
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/goodscenter/model"
	"net/http"
)

// Find 查找商品
func Find(ctx *crpc.Context) {
	good := &model.Goods{
		Id:   200001,
		Name: "生椰拿铁",
	}
	ctx.JSON(http.StatusOK, model.SuccessResponse(good))
}
