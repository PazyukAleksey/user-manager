package cash

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	i, err := strconv.Atoi(os.Getenv("redis_db"))
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("redis_addr"),
		Password: os.Getenv("redis_pass"),
		DB:       i,
	})

	return client, nil
}
