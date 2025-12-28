package infrastructure

import (
	"context"
	"fmt"
	"strconv"
	"tridorian-ztna/pkg/utils"

	"github.com/redis/go-redis/v9"
)

func SetCache() *redis.Client {

	cacheHost := utils.GetEnv("CACHE_HOST", "localhost")
	cachePort := utils.GetEnv("CACHE_PORT", "6379")
	cachePass := utils.GetEnv("CACHE_PASSWORD", "P@ssw0rd")
	cacheDB := utils.GetEnv("CACHE_DB", "0")

	dbNum, err := strconv.Atoi(cacheDB)
	if err != nil {
		panic(fmt.Sprintf("Invalid CACHE_DB value: %v", err))
	}

	valkeyClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cacheHost, cachePort),
		Password: cachePass,
		DB:       dbNum,
	})

	_, err = valkeyClient.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Valkey: %v", err))
	}

	return valkeyClient
}
