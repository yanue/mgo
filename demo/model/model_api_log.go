package model

import (
	"github.com/yanue/mgo"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

// 接口日志相关mongo操作

type apiLogModel struct {
	Mgo mgo.Mgo
}

func NewApiLogModel(dbName string) *apiLogModel {
	m := new(apiLogModel)
	m.Mgo.SetDbName(dbName)
	m.Mgo.SetCollName("api_log")
	go m.createIndex()
	return m
}

func (model *apiLogModel) createIndex() {
	log.Println("正在创建索引： api_log")
	_, _ = model.Mgo.CreateIndex(bson.D{bson.E{Key: "uid", Value: -1}}, false)
	_, _ = model.Mgo.CreateIndex(bson.D{bson.E{Key: "created", Value: -1}}, false)
	log.Println("创建索引完毕： api_log")
}

// 通过单个字段查找数据
func (model *apiLogModel) GetByField(field string, val interface{}) (item *ApiLog, err error) {
	item = new(ApiLog)
	err = model.Mgo.GetByField(item, field, val)
	return
}

// 通过多个字段map查询单个数据
func (model *apiLogModel) GetOneByMap(where map[string]interface{}, sorts ...map[string]int) (item *ApiLog, err error) {
	item = new(ApiLog)
	err = model.Mgo.GetOneByMap(item, where, sorts...)
	return
}

// 通过多个字段map查询多条数据
func (model *apiLogModel) GetAllByMap(where map[string]interface{}, sorts ...map[string]int) (items []*ApiLog, err error) {
	items = make([]*ApiLog, 0)
	err = model.Mgo.GetAllByMap(&items, where, sorts...)
	return
}

// 根据id获取订单信息
func (model *apiLogModel) Get(id int) (*ApiLog, error) {
	return model.GetByField("_id", id)
}

// 获取订单列表信息 - page从1开始
func (model *apiLogModel) List(where map[string]interface{}, page, size int, sorts ...map[string]int) (items []*ApiLog, err error) {
	items = make([]*ApiLog, 0)
	err = model.Mgo.List(&items, where, page, size, sorts...)
	return
}

// 创建
func (model *apiLogModel) Create(item *ApiLog) (int, error) {
	// 获取id
	item.Id = autoId.getIncrId(&model.Mgo)
	item.Created = int(time.Now().Unix())
	err := model.Mgo.Create(item)
	return item.Id, err
}

// 更新 - 通过map匹配字段
func (model *apiLogModel) Update(id int, input map[string]interface{}) error {
	err := model.Mgo.Update(id, input)
	return err
}

// 更新 - 通过结构体 (!!注意!! 会以新数据覆盖)
func (model *apiLogModel) Save(data *ApiLog) (err error) {
	return model.Mgo.Save(data.Id, data)
}

func (model *apiLogModel) Delete(id, uid int) error {
	return model.Update(id, map[string]interface{}{"is_deleted": 1})
}
