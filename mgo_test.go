package mgo

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
	"time"
)

func TestMgo_Connect(t *testing.T) {
	_, err := InitMongoClient("mongodb://127.0.0.1:27017", "test", 20)
	if err != nil {
		log.Println("init mongo err: ", err.Error())
		return
	}
	_, _ = TestModel.CreateIndex(bson.D{bson.E{Key: "uid", Value: -1}, bson.E{Key: "account_id", Value: -1}, bson.E{Key: "period", Value: -1}}, false)
	_, _ = TestModel.CreateIndex(bson.D{bson.E{Key: "uid", Value: -1}, bson.E{Key: "period", Value: -1}, bson.E{Key: "account_id", Value: -1}}, false)
	c := TestModel.DropIndex("name_1")
	log.Println("c", c)
	a, b := TestModel.CreateIndex(bson.D{}, true)
	log.Println("CreateIndex a, b", a, b)
	item := &Test{
		Name: "aaaaaaaaa",
		Data: map[string]interface{}{"a": 1, "b": 2},
	}

	id, err := TestModel.Create(item)
	log.Println("id", id, err)

	err = TestModel.Update(id, map[string]interface{}{"age": 18})
	log.Println("update", err)

	row, err := TestModel.Get(id)
	log.Println("row", row, err)

	row.Name = "dddddddddddd"
	err = TestModel.Save(row)
	log.Println("Save", row, err)

	list, err := TestModel.GetAllByMap(map[string]interface{}{"name": "ccccc", "age": 28})
	log.Println("list", list, err)

	// 聚合查询
	//pipeline := `[
	//	{"$match": { "color": "Red" }},
	//	{"$group": { "_id": "$brand", "count": { "$sum": 1 } }},
	//	{"$project": { "brand": "$_id", "_id": 0, "count": 1 }}
	//]`
	pipeline := `[
		{"$match": { "name": "ccccc" }}
	]`
	items, err := TestModel.Aggregate(pipeline)
	fmt.Println("items", items, err)
	for _, v := range items {
		fmt.Println("v", *v)
	}
}

// 多库多表模式
func newTestMultiModel(dbName, collName string) *testModel {
	m := new(testModel)
	m.Mgo.SetDbName(dbName)
	m.Mgo.SetCollName(collName)
	return m
}

// 单例模式对外导出方法
var TestModel = newTestModel()

func newTestModel() *testModel {
	m := new(testModel)
	m.Mgo.SetCollName("test")
	return m
}

// 表(集合)信息
type Test struct {
	Id        int         `json:"id" bson:"_id"` // 自增涨id
	Name      string      `json:"name" bson:"name"`
	Age       int         `json:"age" bson:"age"`
	Data      interface{} `json:"data" bson:"data"`
	IsDeleted int         `json:"is_deleted" json:"is_deleted"`
	Created   int         `json:"created" bson:"created"`
}

// 对mongo表操作处理
type testModel struct {
	Mgo // 需要使用匿名结构,保证初始化不用单独new
}

// 通过单个字段查找数据
func (model *testModel) GetByField(field string, val interface{}) (item *Test, err error) {
	item = new(Test)
	err = model.Mgo.GetByField(item, field, val)
	return
}

// 通过多个字段map查询单个数据
func (model *testModel) GetOneByMap(where map[string]interface{}, sorts ...map[string]int) (item *Test, err error) {
	item = new(Test)
	err = model.Mgo.GetOneByMap(item, where, sorts...)
	return
}

// 通过多个字段map查询多条数据
func (model *testModel) GetAllByMap(where map[string]interface{}, sorts ...map[string]int) (items []*Test, err error) {
	items = make([]*Test, 0)
	err = model.Mgo.GetAllByMap(&items, where, sorts...)
	return
}

// 根据id获取信息
func (model *testModel) Get(id int) (*Test, error) {
	return model.GetByField("_id", id)
}

// 获取列表 - page从1开始
func (model *testModel) List(where map[string]interface{}, page, size int, sorts ...map[string]int) (items []*Test, err error) {
	items = make([]*Test, 0)
	err = model.Mgo.List(&items, where, page, size, sorts...)
	return
}

// 创建
func (model *testModel) Create(item *Test) (int, error) {
	// 自增获取id
	item.Id = model.GetLastId()
	item.Created = int(time.Now().Unix())
	err := model.Mgo.Create(item)
	return item.Id, err
}

// 更新 - 通过map匹配字段
func (model *testModel) Update(id int, input map[string]interface{}) error {
	err := model.Mgo.Update(id, input)
	return err
}

// 更新 - 通过结构体 (!!注意!! 会以新数据覆盖)
func (model *testModel) Save(data *Test) (err error) {
	return model.Mgo.Save(data.Id, data)
}

// 软删
func (model *testModel) Delete(id int) error {
	return model.Update(id, map[string]interface{}{"is_deleted": 1})
}

// 硬删
func (model *testModel) ForceDelete(id int) error {
	return model.Mgo.ForceDelete(id)
}

// 聚合查询
func (model *testModel) Aggregate(pipeStr string) (items []*Test, err error) {
	items = make([]*Test, 0)
	err = model.Mgo.Aggregate(pipeStr, &items)
	return
}
