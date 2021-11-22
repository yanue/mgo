module github.com/yanue/mgo/demo

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/yanue/mgo v0.0.0-20200819014032-c978fba44fa6
	go.mongodb.org/mongo-driver v1.5.3
)

replace github.com/yanue/mgo => ../

go 1.13
