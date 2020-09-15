package main

import (
	"fmt"
	"github.com/yanue/mgo"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	_, err := mgo.InitMongoClient("mongodb://localhost:27017", "test", 20)
	if err != nil {
		log.Println("init mongo err: ", err.Error())
		return
	}
	runTest1()
	runTest2()
}

func runTest1() {
	item := &Test{
		Name: "aaaaaaaaa",
		Data: map[string]interface{}{"a": 1, "b": 2},
	}

	id, err := TestModel.Create(item)
	log.Println("id", id, err)

	row, err := TestModel.Get(id)
	log.Println("row", row, err)

	err = TestModel.Update(id, map[string]interface{}{"name": "bbbbbbbbb"})
	log.Println("update", err)

	list, err := TestModel.GetAllByMap(map[string]interface{}{"name": "bbbbbbbbb"})
	log.Println("list", list, err)

	// 聚合查询
	//pipeline := `[
	//	{"$match": { "color": "Red" }},
	//	{"$group": { "_id": "$brand", "count": { "$sum": 1 } }},
	//	{"$project": { "brand": "$_id", "_id": 0, "count": 1 }}
	//]`
	pipeline := `[
		{"$match": { "name": bbbbbbbbb }}
	]`
	items, err := TestModel.Aggregate(pipeline)
	fmt.Println("items", items, err)
	for _, v := range items {
		fmt.Println("v", *v)
	}
}

func runTest2() {
	item := &Test2{
		Key:    "aaa",
		SubKey: "bbb",
		Data:   map[string]interface{}{"a": 1, "b": 2},
	}

	err := Test2Model.Create(item)
	log.Println("id", err)

	row, err := Test2Model.GetOneByMap(map[string]interface{}{"key": "aaa"})
	log.Println("row", row, err)

	row2, err := Test2Model.GetByMongoFind(map[string]interface{}{"key": "aaa"})
	log.Println("row2", row2, err)
}
