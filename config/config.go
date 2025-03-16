package config

import "os"

var (
	MySQLDSN  = os.Getenv("MYSQL_DSN")
	RedisAddr = os.Getenv("REDIS_ADDR")
)