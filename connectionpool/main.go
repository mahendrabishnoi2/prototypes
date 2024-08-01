package main

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var concurrency = 150
var poolSize = 10

func main() {
	ctx := context.Background()
	dbConfig := NewDbConfig()

	timeIt("with pooling", func() {
		withPooling(ctx, dbConfig, concurrency)
	})
	timeIt("without pooling", func() {
		withoutConnectionPooling(ctx, dbConfig, concurrency)
	})
}

func withPooling(ctx context.Context, dbConfig *DbConfig, concurrency int) {
	pool, err := NewConnectionPool(*dbConfig, poolSize)
	if err != nil {
		slog.Error("failed to create connection pool", slog.Any("err", err))
		return
	}

	var wg = &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go queryPool(ctx, pool, wg)
	}

	wg.Wait()
	pool.Close()
}

func withoutConnectionPooling(ctx context.Context, dbConfig *DbConfig, concurrency int) {
	db, err := openDb(dbConfig)
	if err != nil {
		slog.Error("failed to open db", slog.Any("err", err))
		return
	}

	var wg = &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go openConnectionAndQuery(ctx, db, wg)
	}

	wg.Wait()
}

func openConnectionAndQuery(ctx context.Context, db *sql.DB, wg *sync.WaitGroup) error {
	defer wg.Done()
	conn, err := db.Conn(ctx)
	if err != nil {
		slog.Error("failed to open connection", slog.Any("err", err))
		return err
	}
	defer conn.Close()

	err = query(ctx, conn)
	if err != nil {
		slog.Error("failed to query", slog.Any("err", err))
		return err
	}
	return nil
}

func queryPool(ctx context.Context, pool *ConnectionPool, wg *sync.WaitGroup) error {
	conn := pool.Get()
	defer pool.Put(conn)
	defer wg.Done()
	err := query(ctx, conn)
	if err != nil {
		slog.Error("failed to query", slog.Any("err", err))
		return err
	}
	return nil
}

func query(ctx context.Context, conn *sql.Conn) error {
	res, err := conn.QueryContext(ctx, "SELECT 1")
	if err != nil {
		slog.Error("failed to query", slog.Any("err", err))
		return err
	}
	defer res.Close()

	for res.Next() {
		var one int
		err := res.Scan(&one)
		if err != nil {
			slog.Error("failed to scan", slog.Any("err", err))
			return err
		}
	}
	return nil
}

func openDb(dbConfig *DbConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbConfig.ConnectionString())
	if err != nil {
		return nil, err
	}
	return db, nil
}

func timeIt(name string, f func()) {
	start := time.Now()
	f()
	slog.Info(name, slog.Any("time", time.Since(start)))
}
