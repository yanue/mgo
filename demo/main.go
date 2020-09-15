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
	id, err = model.ApiLogModel.Create(l)
	log.Println("id, err", id, err)
	list2, err := model.ApiLogModel.GetAllByMap(map[string]interface{}{}, map[string]int{})
	log.Println("list,err", list)
	for _, item := range list2 {
		log.Println("api_log", *item)
	}
}
