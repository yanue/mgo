package model

import (
	"github.com/go-redis/redis"
	"github.com/yanue/mgo"
	"log"
	"os"
)

// 初始化
func Init() {
	_, em := mgo.InitMongoClient("mongodb://localhost:27017", "test", 50)
	if em != nil {
		log.Println("mongo连接失败:", em.Error())
		os.Exit(0)
	}
	e1 := NewRedisClient()
	if e1 != nil {
		log.Println("---------------------------------------------------------")
		log.Println("redis连接失败:", e1.Error())
		os.Exit(0)
	}
	InitMongo()
	InitCache()
}

/**
 * 创建索引，增改字段等
 */
func InitMongo() {
	UserModel.createIndex()
	ApiLogModel.createIndex()
}

var redisClient *redis.Client

// 缓存中key
type CacheKey = string

/**
 * 初始化redis连接
 *
 * return: error
 */
func NewRedisClient() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// 通过 cient.Ping() 来检查是否成功连接到了 redis 服务器
	_, err := redisClient.Ping().Result()
	if err == nil {
		log.Println("Connected to Redis!")
	}

	return err
}

/**
 * 初始化缓存
 * -- 需要先NewRedisClient
 */
func InitCache() {
	NewAutoIncreaseId()
}
