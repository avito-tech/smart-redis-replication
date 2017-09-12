// Package main это пример приложения для бесшовной миграции данных
package main

import (
	"flag"
	"log"

	"github.com/avito-tech/smart-redis-replication/backlog"
)

func main() {
	conf := Config{}
	conf.redis.host = flag.String("redis-host", "", "redis master hostname")
	conf.redis.port = flag.Int("redis-port", 0, "redis master port")
	conf.service.address = flag.String("service-address", "", "service address")
	conf.useStderr = flag.Bool("use-stderr", false, "logger useStderr")
	conf.useMetrics = flag.Bool("use-metrics", true, "enable metrics")
	conf.replica.CacheRDB = true
	conf.replica.CacheRDBFile = "/tmp/rdb.cache"
	conf.replica.BacklogSize = backlog.DefaultBacklogSize
	conf.replica.ReadRDB = true
	flag.Parse()

	err := conf.Check()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	migration, err := NewMigration(conf)
	if err != nil {
		log.Fatalf("create migration error: %v", err)
	}
	err = migration.Start()
	if err != nil {
		log.Fatalf("migration error: %v", err)
	}
}
