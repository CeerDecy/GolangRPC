package service

import (
	context "context"
	"github/CeerDecy/RpcFrameWork/goodscenter/api"
	"github/CeerDecy/RpcFrameWork/goodscenter/model"
)

type GoodsService struct {
	api.UnimplementedGoodsApiServer
}

func (g *GoodsService) Find(ctx context.Context, request *api.GoodsRequest) (*api.GoodsResponse, error) {
	goods := &api.Goods{
		Id:   1000,
		Name: "长岛冰茶 - GRPC",
	}
	res := &api.GoodsResponse{
		Code: 200,
		Msg:  "success",
		Data: goods,
	}
	return res, nil
}

func (g *GoodsService) mustEmbedUnimplementedGoodsApiServer() {
	//TODO implement me
	panic("implement me")
}

type GoodsRpcService struct {
}

func (s *GoodsRpcService) Find(id int64) *model.Response {
	goods := model.Goods{Id: id, Name: "RpcGoodsName2"}
	return model.SuccessResponse(goods)
}
