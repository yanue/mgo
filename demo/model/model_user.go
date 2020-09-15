package model

import (
	"github.com/yanue/mgo"
	"log"
	"time"
)

// 用户相关mongo操作
var UserModel = newUserModel()

type userModel struct {
	Mgo mgo.Mgo
}

func newUserModel() *userModel {
	m := new(userModel)
	m.Mgo.SetCollName("user")
	return m
}

func (model *userModel) createIndex() {
	log.Println("正在创建索引： user")
	_, _ = model.Mgo.CreateIndex(map[string]int{"phone": 1}, true)
	_, _ = model.Mgo.CreateIndex(map[string]int{"user_name": 1}, true)
	_, _ = model.Mgo.CreateIndex(map[string]int{"created": -1}, false)
	log.Println("创建索引完毕： user")
}

// 通过字段查找数据
func (model *userModel) GetByField(field string, val interface{}) (item *User, err error) {
	item = new(User)
	err = model.Mgo.GetByField(item, field, val)
	return
}

// 通过多个字段map查询单个数据
func (model *userModel) GetOneByMap(where map[string]interface{}, sorts ...map[string]int) (item *User, err error) {
	item = new(User)
	err = model.Mgo.GetOneByMap(item, where, sorts...)
	return
}

// 通过多个字段map查询多条数据
func (model *userModel) GetAllByMap(where map[string]interface{}, sorts ...map[string]int) (items []*User, err error) {
	items = make([]*User, 0)
	err = model.Mgo.GetAllByMap(&items, where, sorts...)
	return
}

// 通过id查找
func (model *userModel) Get(uid int) (*User, error) {
	return model.GetByField("_id", uid)
}

// 列表
func (model *userModel) List(where map[string]interface{}, page, size int, sorts ...map[string]int) (items []*User, err error) {
	items = make([]*User, 0)
	err = model.Mgo.List(&items, where, page, size, sorts...)
	return
}

// 创建用户
func (model *userModel) Create(item *User) (int, error) {
	// 获取id
	item.Id = autoId.getIncrId(&model.Mgo)
	item.Created = int(time.Now().Unix())
	err := model.Mgo.Create(item)
	return item.Id, err
}

// 更新用户
func (model *userModel) Update(id int, input map[string]interface{}) error {
	return model.Mgo.Update(id, input)
}

// 更新 - 通过结构体 (!!注意!! 会以新数据覆盖)
func (model *userModel) Save(data *User) (err error) {
	return model.Mgo.Save(data.Id, data)
}
