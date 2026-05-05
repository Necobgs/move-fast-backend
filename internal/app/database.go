package app

import (
	"context"

	"github.com/Necobgs/move-fast-backend/configs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis(ctx *context.Context) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     configs.Cfg.RdHost + ":" + configs.Cfg.RdPort,
		Password: configs.Cfg.RdPassword,
	})

	if rdb.Ping(*ctx).Err() != nil {
		panic("cant connect do redis")
	}
	return rdb
}

func ConnectDatabase(ctx context.Context, url string) *pgxpool.Pool {
	dbConn, err := pgxpool.New(ctx, url)
	if err != nil {
		panic(err)
	}

	err = dbConn.Ping(ctx)
	if err != nil {
		panic(err)
	}

	return dbConn
}
