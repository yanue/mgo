package main

import (
	"github.com/yanue/mgo"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// 单例模式对外导出方法
var TestModel = new(testModel)

// 表(集合)信息
type Test struct {
	Id        int         `json:"id" bson:"_id"` // 自增涨id
	Name      string      `json:"name" bson:"name"`
	Data      interface{} `json:"data" bson:"data"`
	IsDeleted int         `json:"is_deleted" json:"is_deleted"`
	Created   int         `json:"created" bson:"created"`
}

// 对mongo表操作处理
// 主要目的是mongo表连接信息
type testModel struct {
	mgo.Mgo // 需要使用匿名结构,保证初始化不要单独new
	lastId  int
	coll    *mongo.Collection
}

// 连接Collection
func (model *testModel) getCollection() *mongo.Collection {
	if nil == model.coll {
		// 重点: 表名(集合名)
		model.coll = model.Mgo.GetCollection("test")
	}
	return model.coll
}

// 获取自增id (从1开始)
func (model *testModel) GetLastId() int {
	return model.Mgo.GetLastId(model.getCollection()) + 1
}

// 统计
func (model *testModel) Count(where map[string]interface{}) (cnt int64, err error) {
	return model.Mgo.Count(model.getCollection(), where)
}

// 通过单个字段查找数据
func (model *testModel) GetByField(field string, val interface{}) (item *Test, err error) {
	item = new(Test)
	err = model.Mgo.GetByField(model.getCollection(), item, field, val)
	return
}

// 通过多个字段map查询单个数据
func (model *testModel) GetOneByMap(where map[string]interface{}, sorts ...map[string]int) (item *Test, err error) {
	item = new(Test)
	err = model.Mgo.GetOneByMap(model.getCollection(), item, where, sorts...)
	return
}

// 通过多个字段map查询多条数据
func (model *testModel) GetAllByMap(where map[string]interface{}, sorts ...map[string]int) (items []*Test, err error) {
	items = make([]*Test, 0)
	err = model.Mgo.GetAllByMap(model.getCollection(), &items, where, sorts...)
	return
}

// 根据id获取信息
func (model *testModel) Get(id int) (*Test, error) {
	return model.GetByField("_id", id)
}

// 获取列表 - page从1开始
func (model *testModel) List(where map[string]interface{}, page, size int, sorts ...map[string]int) (items []*Test, err error) {
	items = make([]*Test, 0)
	err = model.Mgo.List(model.getCollection(), &items, where, page, size, sorts...)
	return
}

// 创建
func (model *testModel) Create(item *Test) (int, error) {
	// 自增获取id
	item.Id = model.GetLastId() + 1
	item.Created = int(time.Now().Unix())
	err := model.Mgo.Create(model.getCollection(), item)
	return item.Id, err
}

// 更新 - 通过map匹配字段
func (model *testModel) Update(id int, input map[string]interface{}) error {
	err := model.Mgo.Update(model.getCollection(), id, input)
	return err
}

// 更新 - 通过map匹配字段
func (model *testModel) UpdateByMap(where map[string]interface{}, input map[string]interface{}) error {
	err := model.Mgo.UpdateByMap(model.getCollection(), where, input)
	return err
}

// 更新 - 通过结构体 (!!注意!! 会以新数据覆盖)
func (model *testModel) Save(data *Test) (err error) {
	return model.Mgo.Save(model.getCollection(), data.Id, data)
}

// 软删
func (model *testModel) Delete(id int) error {
	return model.Update(id, map[string]interface{}{"is_deleted": 1})
}

// 硬删
func (model *testModel) ForceDelete(id int) error {
	return model.Mgo.ForceDelete(model.getCollection(), id)
}

// 硬删
func (model *testModel) ForceDeleteByMap(where map[string]interface{}) error {
	return model.Mgo.ForceDeleteByMap(model.getCollection(), where)
}

// 聚合查询
func (model *testModel) Aggregate(pipeStr string) (items []*Test, err error) {
	items = make([]*Test, 0)
	err = model.Mgo.Aggregate(model.getCollection(), pipeStr, items)
	return
}
