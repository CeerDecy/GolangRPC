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

func TestSaveUserBatch(t *testing.T) {
	users := []any{
		&User{Username: "Ceer", Password: "12423425", Age: 21},
		&User{Username: "Decy", Password: "gmsj84h", Age: 19},
	}
	SaveUserBatch(users)
}

func TestUpdate(t *testing.T) {
	Update()
}
