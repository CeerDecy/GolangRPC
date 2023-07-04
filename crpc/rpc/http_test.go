package rpc

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	fmt.Println(toValues(map[string]any{
		"id":   1001,
		"name": "长岛冰茶",
	}))
}
