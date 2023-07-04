package service

import "github/CeerDecy/RpcFrameWork/crpc/rpc"

type GoodsService struct {
	Find func(args map[string]any) ([]byte, error) `crpc:"GET,/goods/find"`
}

func (g *GoodsService) Evn() *rpc.HttpConfig {
	return &rpc.HttpConfig{
		Host: "127.0.0.1",
		Port: 8001,
	}
}
