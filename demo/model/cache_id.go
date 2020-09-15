package model

import (
	"github.com/go-redis/redis"
	"github.com/yanue/mgo"
	"log"
)

// 自增id生成器-支持分布式
// 用于mongodb的_id字段自增

var autoId *AutoIncreaseId

type AutoIncreaseId struct {
	redis         *redis.Client
	autoIdBaseKey string
}

func NewAutoIncreaseId() {
	// 初始化自增id
	autoId = new(AutoIncreaseId)
	// redis前缀key
	autoId.autoIdBaseKey = "auto_id:"
	autoId.redis = redisClient
	// 从mongo同步最大id
	autoId.syncIdFromDb()
}

/**
 * 从mongo同步最新id
 */
func (a *AutoIncreaseId) syncIdFromDb() {
	// 需要同步最新id列表
	var list = []mgo.IMgo{
		&UserModel.Mgo,
	}
	for _, table := range list {
		a.redis.Set(a.autoIdBaseKey+table.GetCollName(), table.GetLastId(), 0)
	}
}

// 通用获取自增id
func (a *AutoIncreaseId) getIncrId(table mgo.IMgo) int {
	key := a.autoIdBaseKey + table.GetCollName()
	// 通过redis自增id(分布式)
	lastId, err := a.redis.Incr(key).Result()
	if err != nil {
		lastId = int64(table.GetLastId())
		log.Println("getIncrId", "table", table.GetCollName(), "lastId", lastId, "err", err)
	}
	// 更新自身属性,避免下次
	table.SetLastId(int(lastId))
	return int(lastId)
}
