package main

import (
	"fmt"
	"github.com/yanue/mgo/demo/model"
	"log"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	model.Init()
}

func main() {
	tmpStr := fmt.Sprintf("%d", time.Now().Unix())
	// 创建用户
	u := new(model.User)
	u.Phone = tmpStr
	u.UserName = tmpStr
	id, err := model.UserModel.Create(u)
	log.Println("id, err", id, err)
	list, err := model.UserModel.GetAllByMap(map[string]interface{}{}, map[string]int{})
	log.Println("list,err", list)
	for _, item := range list {
		log.Println("item", *item)
	}
	// 创建记录
	l := new(model.ApiLog)
	l.ApiName = "创建用户"
	// 不同库
	m := model.NewApiLogModel("test_admin")
	id, err = m.Create(l)
	log.Println("id, err", id, err)
	list2, err := m.GetAllByMap(map[string]interface{}{}, map[string]int{})
	log.Println("list,err", list)
	for _, item := range list2 {
		log.Println("api_log", *item)
	}

	// 不同库
	m1 := model.NewApiLogModel("test_user")
	id, err = m1.Create(l)
	log.Println("id, err", id, err)
	list2, err = m1.GetAllByMap(map[string]interface{}{}, map[string]int{})
	log.Println("list,err", list)
	for _, item := range list2 {
		log.Println("api_log", *item)
	}
}
