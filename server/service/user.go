package service

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github/CeerDecy/RpcFrameWork/crpc/orm"
	"net/url"
)

type User struct {
	Id       int64  `corm:"id,auto_increment"`
	Username string `corm:"username"`
	Password string `corm:"password"`
	Age      int    `corm:"age"`
}

func SaveUser(user *User) {
	source := fmt.Sprintf("root:174878@tcp(nullpoint.com.cn:3306)/crpc?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", source)
	id, _, err := db.NewSession().Table("user").Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	err = db.Close()
	if err != nil {
		panic(err)
	}
}

func SaveUserBatch(users []any) {
	source := fmt.Sprintf("root:174878@tcp(nullpoint.com.cn:3306)/crpc?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", source)
	id, _, err := db.NewSession().Table("user").InsertBatch(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	err = db.Close()
	if err != nil {
		panic(err)
	}
}

func Update() {
	source := fmt.Sprintf("root:174878@tcp(nullpoint.com.cn:3306)/crpc?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", source)
	user := &User{
		Username: "手写的从前",
		Password: "231412312",
		Age:      23,
	}
	id, _, err := db.NewSession().Table("user").Where("id", 10000).Update(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	err = db.Close()
	if err != nil {
		panic(err)
	}
}

func SelectOne() {
	source := fmt.Sprintf("root:174878@tcp(nullpoint.com.cn:3306)/crpc?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", source)
	user := new(User)
	err := db.NewSession().
		Table("user").
		Where("id", 10000).
		SelectOne(user)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", user)
}

func Count() {
	source := fmt.Sprintf("root:174878@tcp(nullpoint.com.cn:3306)/crpc?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", source)
	count, err := db.NewSession().Table("user").Count()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}
