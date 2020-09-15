package mgo

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
	"time"
)

type IMgo interface {
	SetCollName(collName string)
	GetCollName() string
	GetCollection() *mongo.Collection
	GetLastId() int
	SetLastId(int)
	Count(where map[string]interface{}) (cnt int64, err error)
	Update(id interface{}, input map[string]interface{}) error
	UpdateByMap(where map[string]interface{}, input map[string]interface{}) error
	Create(item interface{}) error
	Save(id interface{}, data interface{}) error
	ForceDelete(id interface{}) error
	ForceDeleteByMap(where map[string]interface{}) error
	GetByField(result interface{}, field string, val interface{}) (err error)
	GetOneByMap(result interface{}, where map[string]interface{}, sorts ...map[string]int) (err error)
	GetAllByMap(results interface{}, where map[string]interface{}, sorts ...map[string]int) (err error)
	List(results interface{}, where map[string]interface{}, page, size int, sorts ...map[string]int) (err error)
	Aggregate(pipeStr string, results interface{}) error
}

var mgoCli *mongo.Client
var mongoDbName string

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func InitMongoClient(mongoUri string, dbName string, maxPoolSize uint64) (cli *mongo.Client, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cli, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri).SetMaxPoolSize(maxPoolSize)) // 最大连接池
	if err != nil {
		log.Println("mongo.Connect err:", err.Error())
		return
	}
	// Check the connection
	err = cli.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("mongo.Ping err", err.Error())
		return
	}
	// mongo连接信息
	mgoCli = cli
	// mongo库名
	mongoDbName = dbName
	log.Println("Connected to MongoDB!")
	return
}

// 所有model结构体继承Mgo
type Mgo struct {
	IMgo
	coll     *mongo.Collection
	collName string
	lastId   int
}

// 连接表(集合collection)
func (model *Mgo) SetCollName(collName string) {
	model.collName = collName
}

// 连接表(集合collection)
func (model *Mgo) GetCollName() string {
	return model.collName
}

// 连接表(集合collection)
func (model *Mgo) GetCollection() *mongo.Collection {
	if len(model.collName) == 0 {
		panic("请使用SetCollName方法设置collName")
	}
	if nil == model.coll {
		// 重点: 表名(集合名)
		model.coll = mgoCli.Database(mongoDbName).Collection(model.collName)
	}
	return model.coll
}

// 获取自增id
func (model *Mgo) GetLastId() int {
	if model.lastId == 0 {
		model.lastId = model.getLastId()
	}
	return model.lastId
}

// 设置自增id
func (model *Mgo) SetLastId(lastId int) {
	model.lastId = lastId
}

/**
 * 获取自增_id最大值
 *
 * param: *mongo.Collection coll
 * return: int
 */
func (model *Mgo) getLastId() int {
	// Sort by `_id` field descending
	result := model.GetCollection().FindOne(getContext(), bson.M{}, options.FindOne().SetSort(bson.D{{"_id", -1}}))
	if err := result.Err(); err != nil {
		log.Println("getLastId err:", err.Error())
		return 0
	}
	resp := &struct {
		Id int `json:"id" bson:"_id"`
	}{}
	err := result.Decode(resp)
	if err != nil {
		log.Println("getLastId err", err.Error())
		return 0
	}
	return resp.Id
}

/**
 * 统计数据
 *
 * param: *mongo.Collection      coll
 * param: map[string]interface{} where
 * return: int64
 * return: error
 */
func (model *Mgo) Count(where map[string]interface{}) (cnt int64, err error) {
	// 组装数据
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	cnt, err = model.GetCollection().CountDocuments(context.Background(), filter)
	return
}

/**
 * 通过单个字段查找数据
 *
 * param: *mongo.Collection coll
 * param: interface{}       result - 结构体指针
 * param: string            field
 * param: interface{}       val
 * return: error
 */
func (model *Mgo) GetByField(result interface{}, field string, val interface{}) (err error) {
	res := model.GetCollection().FindOne(context.Background(), bson.M{field: val})
	if err = res.Err(); err != nil {
		// todo 没数据的区分
		return
	}
	err = res.Decode(result)
	return
}

/**
 * 通过多个字段map查询单个数据
 *
 * param: *mongo.Collection      coll
 * param: interface{}            result - 结构体指针
 * param: map[string]interface{} where
 * param: ...map[string]int      sorts
 * return: error
 */
func (model *Mgo) GetOneByMap(result interface{}, where map[string]interface{}, sorts ...map[string]int) (err error) {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	opts := options.FindOne()
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	res := model.GetCollection().FindOne(context.Background(), filter, opts)
	if err = res.Err(); err != nil {
		return
	}
	err = res.Decode(result)
	return
}

/**
 * 通过多个字段map查询多条数据

 * param: *mongo.Collection      coll
 * param: interface{}            results - map的指针
 * param: map[string]interface{} where
 * param: ...map[string]int      sorts
 * return: error
 */
func (model *Mgo) GetAllByMap(results interface{}, where map[string]interface{}, sorts ...map[string]int) (err error) {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	opts := options.Find()
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	ctx := context.Background()
	cur, err := model.GetCollection().Find(ctx, filter, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	// 解析到map
	err = cur.All(ctx, results)
	return
}

/**
 * 通过多个字段map查询多条数据

 * param: *mongo.Collection      coll
 * param: interface{}            results - map的指针
 * param: int      				 page - 页码(从1开始)
 * param: int                    size
 * param: map[string]interface{} where
 * param: ...map[string]int      sorts
 * return: error
 */
func (model *Mgo) List(results interface{}, where map[string]interface{}, page, size int, sorts ...map[string]int) (err error) {
	// page从1开始
	if page > 1 {
		page--
	} else {
		page = 0
	}
	// 组装数据
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	opts := options.Find()
	opts.SetLimit(int64(size))
	opts.SetSkip(int64(page * size))
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	var ctx = context.Background()
	cur, err := model.GetCollection().Find(ctx, filter, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	// 解析结果
	err = cur.All(ctx, results)
	return
}

/**
 * 更新数据 - 通过map匹配字段
 *
 * param: *mongo.Collection      coll
 * param: int                    id
 * param: map[string]interface{} input
 * return: error
 */
func (model *Mgo) Update(id interface{}, input map[string]interface{}) error {
	// 组装数据
	data := bson.M{}
	for k, v := range input {
		data[k] = v
	}
	_, err := model.GetCollection().UpdateOne(getContext(), bson.M{"_id": id}, bson.D{{"$set", data}})
	return err
}

/**
 * 更新数据 - 通过map匹配字段
 *
 * param: *mongo.Collection      coll
 * param: int                    id
 * param: map[string]interface{} input
 * return: error
 */
func (model *Mgo) UpdateByMap(where map[string]interface{}, input map[string]interface{}) error {
	// 组装数据
	whereC := bson.M{}
	for k, v := range where {
		whereC[k] = v
	}
	// 组装数据
	data := bson.M{}
	for k, v := range input {
		data[k] = v
	}
	_, err := model.GetCollection().UpdateMany(getContext(), whereC, bson.D{{"$set", data}})
	return err
}

/**
 * 插入数据 - 通过结构体
 * param: *mongo.Collection      coll
 * param: map[string]interface{} item
 * return: error
 */
func (model *Mgo) Create(item interface{}) error {
	_, err := model.GetCollection().InsertOne(getContext(), item)
	return err
}

/**
 * 更新数据 - 通过结构体
 * -- !!!注意!!!
 * -- 这里会覆盖所有字段,除了id
 *
 * param: *mongo.Collection      coll
 * param: int                    id
 * param: map[string]interface{} input
 * return: error
 */
func (model *Mgo) Save(id interface{}, data interface{}) error {
	_, err := model.GetCollection().UpdateOne(getContext(), bson.M{"_id": id}, bson.D{{"$set", data}})
	return err
}

/**
 * 硬删除一条
 *
 * param: *mongo.Collection      coll
 * param: interface{}            id
 * return: error
 */
func (model *Mgo) ForceDelete(id interface{}) error {
	_, err := model.GetCollection().DeleteOne(getContext(), bson.M{"_id": id})
	return err
}

/**
 * 硬删除多条
 *
 * param: *mongo.Collection      coll
 * param: map[string]interface{} where
 * return: error
 */
func (model *Mgo) ForceDeleteByMap(where map[string]interface{}) error {
	// 组装数据
	whereC := bson.M{}
	for k, v := range where {
		whereC[k] = v
	}
	log.Println("where", whereC)
	_, err := model.GetCollection().DeleteMany(getContext(), whereC)
	return err
}

/**
 * 聚合查询 - aggregate
 *
 * param: *mongo.Collection coll
 * param: string  pipeStr - aggregate操作json字符串
   如:	pipeline := `[
		{"$match": { "color": "Red" }},
		{"$group": { "_id": "$brand", "count": { "$sum": 1 } }},
		{"$project": { "brand": "$_id", "_id": 0, "count": 1 }}
	]`
 * param: interface{}       results
 * return: error
*/
func (model *Mgo) Aggregate(pipeStr string, results interface{}) error {
	var ctx = context.Background()
	opts := options.Aggregate()
	pipe, err := model.parsePipeline(pipeStr)
	if err != nil {
		return err
	}
	//util.Log.Info("pipe", pipeStr, pipe)
	cur, err := model.GetCollection().Aggregate(context.Background(), pipe, opts)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	// 解析结果
	err = cur.All(ctx, results)
	return nil
}

/**
 * 将mongo aggregate操作bson字符串 转换成 mongo.Pipeline
 *
 * param: string str - bson字符串
 * return: mongo.Pipeline
 */
func (model *Mgo) parsePipeline(str string) (pipeline mongo.Pipeline, err error) {
	pipeline = []bson.D{}
	str = strings.TrimSpace(str)
	if strings.Index(str, "[") != 0 {
		var doc bson.M
		if err = json.Unmarshal([]byte(str), &doc); err != nil {
			return
		}
		var v bson.D
		b, err1 := bson.Marshal(doc)
		if err1 != nil {
			err = err1
			return
		}
		err = bson.Unmarshal(b, &v)
		if err != nil {
			return
		}
		pipeline = append(pipeline, v)
	} else {
		var docs []bson.M
		err = json.Unmarshal([]byte(str), &docs)
		if err != nil {
			return
		}
		for _, doc := range docs {
			var v bson.D
			b, err1 := bson.Marshal(doc)
			if err1 != nil {
				err = err1
				return
			}
			err = bson.Unmarshal(b, &v)
			if err != nil {
				return
			}
			pipeline = append(pipeline, v)
		}
	}
	return
}

// 创建索引: keys: map[字段名]排序方式(-1|1)
func (model *Mgo) CreateIndex(keys map[string]int, Unique bool) (string, error) {
	if len(keys) == 0 {
		return "", nil
	}
	idx := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(Unique),
	}
	result, err := model.GetCollection().Indexes().CreateOne(getContext(), idx)
	return result, err
}

// 删除索引
func (model *Mgo) DropIndex(name string) error {
	_, err := model.GetCollection().Indexes().DropOne(getContext(), name)
	return err
}

func getContext() context.Context {
	return context.Background()
}
