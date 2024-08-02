package mgo

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
	"sync"
	"time"
)

type IMgo interface {
	GetMgoCli() *mongo.Client
	SetDbName(dbName string)
	GetDbName() string
	SetCollName(collName string)
	GetCollName() string
	GetCollection() *mongo.Collection
	GetLastId() int
	SetLastId(int)
	Count(where map[string]any) (cnt int64, err error)
	Update(id any, input map[string]any) error
	UpdateByMap(where map[string]any, input map[string]any) error
	Create(item any) error
	Save(id any, item any) error
	InsertMany(item []any) error
	ForceDelete(id any) error
	ForceDeleteByMap(where map[string]any) error
	GetByField(result any, field string, val any) (err error)
	GetAllWithFields(results any, where map[string]any, _sort map[string]int, fields []string) (err error)
	GetOneByMap(result any, where map[string]any, sorts ...map[string]int) (err error)
	GetAllByMap(results any, where map[string]any, sorts ...map[string]int) (err error)
	List(results any, where map[string]any, page, size int, sorts ...map[string]int) (err error)
	ListWithFields(results any, where map[string]any, page, size int, _sort map[string]int, fields []string) (err error)
	Aggregate(pipeStr string, results any) error
	CreateIndex(keys bson.D, Unique bool) (string, error)
	DropIndex(name string) error
}

type dbCollName struct {
	dbName   string
	collName string
}

type Session struct {
	sessions map[dbCollName]*mongo.Collection
	lock     sync.RWMutex
}

var session = newCollectionHub()

func newCollectionHub() *Session {
	s := new(Session)
	s.sessions = make(map[dbCollName]*mongo.Collection, 0)
	return s
}

func (s *Session) getSession(name dbCollName) *mongo.Collection {
	s.lock.Lock()
	coll, ok := s.sessions[name]
	if !ok {
		coll = mgoCli.Database(name.dbName).Collection(name.collName)
		s.sessions[name] = coll
	}
	s.lock.Unlock()
	return coll
}

var mgoCli *mongo.Client
var mgoDefaultDbName string

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func InitMongoClient(mongoUri string, defaultDbName string, maxPoolSize uint64) (cli *mongo.Client, err error) {
	mgoDefaultDbName = defaultDbName
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
	log.Println("Connected to MongoDB!")
	return
}

// 所有model结构体继承Mgo

type Mgo struct {
	IMgo
	session *mongo.Collection
	coll    dbCollName
	lastId  int
}

// 库名一般需要加载conf文件才能确认,需要单独设置

func (s *Mgo) SetDbName(dbName string) {
	s.coll.dbName = dbName
}

// collection根据struct,可以直接init使用

func (s *Mgo) SetCollName(collName string) {
	s.coll.collName = collName
}

func (s *Mgo) GetMgoCli() *mongo.Client {
	return mgoCli
}

func (s *Mgo) GetCollName() string {
	return s.coll.collName
}

func (s *Mgo) GetDbName() string {
	return s.coll.dbName
}

func (s *Mgo) GetCollection() *mongo.Collection {
	if s.session != nil {
		return s.session
	}
	if len(s.coll.collName) == 0 {
		panic("请使用SetDbColl方法设置dbName及collName")
	}
	// 默认库
	if len(s.coll.dbName) == 0 {
		s.coll.dbName = mgoDefaultDbName
	}
	// 重点: 表名(集合名)
	s.session = session.getSession(s.coll)
	return s.session
}

// 获取自增id

func (s *Mgo) GetLastId() int {
	if s.lastId == 0 {
		s.lastId = s.getLastId()
	}
	s.lastId++ // 自增
	return s.lastId
}

// 设置自增id

func (s *Mgo) SetLastId(lastId int) {
	s.lastId = lastId
}

// 获取自增_id最大值

func (s *Mgo) getLastId() int {
	// Sort by `_id` field descending
	result := s.GetCollection().FindOne(getContext(), bson.M{}, options.FindOne().SetSort(bson.D{{"_id", -1}}))
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

// 统计数据

func (s *Mgo) Count(where map[string]any) (cnt int64, err error) {
	cnt, err = s.GetCollection().CountDocuments(context.Background(), where)
	return
}

// 通过单个字段查找数据

func (s *Mgo) GetByField(result any, field string, val any) (err error) {
	res := s.GetCollection().FindOne(context.Background(), bson.M{field: val})
	if err = res.Err(); err != nil {
		return
	}
	err = res.Decode(result)
	return
}

// 通过多个字段map查询单个数据

func (s *Mgo) GetOneByMap(result any, where map[string]any, sorts ...map[string]int) (err error) {
	opts := options.FindOne()
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	res := s.GetCollection().FindOne(context.Background(), where, opts)
	if err = res.Err(); err != nil {
		return
	}
	err = res.Decode(result)
	return
}

// 通过多个字段map查询多条数据

func (s *Mgo) GetAllByMap(results any, where map[string]any, sorts ...map[string]int) (err error) {
	opts := options.Find()
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	ctx := context.Background()
	cur, err := s.GetCollection().Find(ctx, where, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	// 解析到map
	err = cur.All(ctx, results)
	return
}

func (s *Mgo) GetAllWithFields(results any, where map[string]any, _sort map[string]int, fields []string) (err error) {
	opts := options.Find()
	opts.SetSort(_sort)
	if len(fields) > 0 {
		var projection = make(bson.M, 0)
		for _, s2 := range fields {
			projection[s2] = 1
		}
		opts.SetProjection(projection)
	}
	ctx := context.Background()
	cur, err := s.GetCollection().Find(ctx, where, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	// 解析到map
	err = cur.All(ctx, results)
	return
}

// 通过多个字段map查询多条数据

func (s *Mgo) List(results any, where map[string]any, page, size int, sorts ...map[string]int) (err error) {
	// page从1开始
	if page > 1 {
		page--
	} else {
		page = 0
	}
	opts := options.Find()
	opts.SetLimit(int64(size))
	opts.SetSkip(int64(page * size))
	if len(sorts) > 0 {
		opts.SetSort(sorts[0])
	}
	var ctx = context.Background()
	cur, err := s.GetCollection().Find(ctx, where, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	// 解析结果
	err = cur.All(ctx, results)
	return
}

func (s *Mgo) ListWithFields(results any, where map[string]any, page, size int, _sort map[string]int, fields []string) (err error) {
	// page从1开始
	if page > 1 {
		page--
	} else {
		page = 0
	}
	opts := options.Find()
	opts.SetLimit(int64(size))
	opts.SetSkip(int64(page * size))
	opts.SetSort(_sort)
	if len(fields) > 0 {
		var projection = make(bson.M, 0)
		for _, s2 := range fields {
			projection[s2] = 1
		}
		opts.SetProjection(projection)
	}
	var ctx = context.Background()
	cur, err := s.GetCollection().Find(ctx, where, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	// 解析结果
	err = cur.All(ctx, results)
	return
}

// 更新数据 - 通过map匹配字段

func (s *Mgo) Update(id any, input map[string]any) error {
	_, err := s.GetCollection().UpdateOne(getContext(), bson.M{"_id": id}, bson.D{{"$set", input}})
	return err
}

/**
 * 更新数据 - 通过结构体
 * -- !!!注意!!!
 * -- 这里会覆盖所有字段,除了id
 */

func (s *Mgo) Save(id any, data any) error {
	update := make(bson.M, 0)
	// 先解析成bson
	bytes, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	// 解析bson到bson.M
	err = bson.Unmarshal(bytes, &update)
	if err != nil {
		return err
	}
	// 删除id
	delete(update, "_id")
	return s.Update(id, update)
}

// 更新数据 - 通过map匹配字段

func (s *Mgo) UpdateByMap(where map[string]any, input map[string]any) error {
	_, err := s.GetCollection().UpdateMany(getContext(), where, bson.D{{"$set", input}})
	return err
}

// 插入数据 - 通过结构体或map

func (s *Mgo) Create(item any) error {
	_, err := s.GetCollection().InsertOne(getContext(), item)
	return err
}

// 批量插入数据 - 通过结构体或map

func (s *Mgo) InsertMany(items []any) error {
	if len(items) == 0 {
		return nil
	}
	_, err := s.GetCollection().InsertMany(getContext(), items)
	return err
}

// 硬删除一条

func (s *Mgo) ForceDelete(id any) error {
	_, err := s.GetCollection().DeleteOne(getContext(), bson.M{"_id": id})
	return err
}

// ForceDeleteByMap 硬删除多条
func (s *Mgo) ForceDeleteByMap(where map[string]any) error {
	// 组装数据
	whereC := bson.M{}
	for k, v := range where {
		whereC[k] = v
	}
	_, err := s.GetCollection().DeleteMany(getContext(), whereC)
	return err
}

// Aggregate 聚合查询 - aggregate
/*
pipeStr - aggregate操作json字符串
   如:	pipeline := `[
		{"$match": { "color": "Red" }},
		{"$group": { "_id": "$brand", "count": { "$sum": 1 } }},
		{"$project": { "brand": "$_id", "_id": 0, "count": 1 }}
	]`
*/

func (s *Mgo) Aggregate(pipeStr string, results any) error {
	var ctx = context.Background()
	opts := options.Aggregate()
	pipe, err := s.parsePipeline(pipeStr)
	if err != nil {
		return err
	}
	cur, err := s.GetCollection().Aggregate(context.Background(), pipe, opts)
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
func (s *Mgo) parsePipeline(str string) (pipeline mongo.Pipeline, err error) {
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

func (s *Mgo) CreateIndex(keysD bson.D, Unique bool) (string, error) {
	if len(keysD) == 0 {
		return "", nil
	}
	idx := mongo.IndexModel{
		Keys:    keysD,
		Options: options.Index().SetUnique(Unique),
	}
	var keys []string
	for _, e := range keysD {
		keys = append(keys, fmt.Sprintf("%v_%d", e.Key, e.Value))
	}
	key := strings.Join(keys, "_")
	list, err := s.GetCollection().Indexes().ListSpecifications(getContext())
	for _, item := range list {
		if item.Name == key {
			return key, nil
		}
	}
	result, err := s.GetCollection().Indexes().CreateOne(getContext(), idx)
	return result, err
}

// 获取索引列表

func (s *Mgo) ListIndex() ([]*mongo.IndexSpecification, error) {
	return s.GetCollection().Indexes().ListSpecifications(getContext())
}

// 删除索引

func (s *Mgo) DropIndex(name string) error {
	_, err := s.GetCollection().Indexes().DropOne(getContext(), name)
	return err
}

func getContext() context.Context {
	return context.Background()
}
