package config

import "os"

var (
	MYSQL_DSN  = os.Getenv("MYSQL_DSN")
	REDIS_ADDR = os.Getenv("REDIS_ADDR")
)
