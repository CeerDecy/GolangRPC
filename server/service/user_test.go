package service

import "testing"

func TestSaveUser(t *testing.T) {
	user := &User{
		Username: "猛喝威士忌",
		Password: "123456",
		Age:      22,
	}
	SaveUser(user)
}
