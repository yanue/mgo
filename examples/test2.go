package main

import (
	"context"
	"github.com/yanue/mgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// 单例模式对外导出方法
var Test2Model = new(test2Model)

// 表(集合)信息
type Test2 struct {
	Id        *primitive.ObjectID `json:"-" bson:"_id,omitempty"` // 使用mongo的内部objectId
	Key       string              `json:"key" bson:"key"`
	SubKey    string              `json:"sub_key" bson:"sub_key"`
	Data      interface{}         `json:"data" bson:"data"`
	IsDeleted int                 `json:"is_deleted" json:"is_deleted"`
	Created   int                 `json:"created" bson:"created"`
}

// 对mongo表操作处理
// 主要目的是mongo表连接信息
type test2Model struct {
	mgo.Mgo // 需要使用匿名结构,保证初始化不要单独new
	coll    *mongo.Collection
}

// 连接Collection
func (model *test2Model) getCollection() *mongo.Collection {
	if nil == model.coll {
		// 重点: 表名(集合名collection)
		model.coll = model.Mgo.GetCollection("test2")
	}
	return model.coll
}

// 通过多个字段map查询单个数据
func (model *test2Model) GetOneByMap(where map[string]interface{}, sorts ...map[string]int) (item *Test2, err error) {
	item = new(Test2)
	err = model.Mgo.GetOneByMap(model.getCollection(), item, where, sorts...)
	return
}

// 通过多个字段map查询多条数据
func (model *test2Model) GetAllByMap(where map[string]interface{}, sorts ...map[string]int) (items []*Test2, err error) {
	items = make([]*Test2, 0)
	err = model.Mgo.GetAllByMap(model.getCollection(), &items, where, sorts...)
	return
}

// 创建
func (model *test2Model) Create(item *Test2) error {
	item.Created = int(time.Now().Unix())
	err := model.Mgo.Create(model.getCollection(), item)
	return err
}

// 通过mongo.driver原始操作
func (model *test2Model) GetByMongoFind(where map[string]interface{}, sorts ...map[string]int) (item *Test2, err error) {
	coll := model.getCollection()
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	opts := options.FindOne()
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	item = new(Test2)
	res := coll.FindOne(context.Background(), filter, opts)
	if err = res.Err(); err != nil {
		return
	}
	err = res.Decode(item)
	return
}
