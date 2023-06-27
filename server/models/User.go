package models

type User struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age" validate:"required,max=50,min=18"`
}
