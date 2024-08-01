package main

import (
	"context"
	"database/sql"
)

type ConnectionPool struct {
	dbConfig DbConfig

	maxConn int

	// store open connections
	conns chan *sql.Conn
}

func NewConnectionPool(dbConfig DbConfig, maxConn int) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		dbConfig: dbConfig,
		maxConn:  maxConn,
		conns:    make(chan *sql.Conn, maxConn),
	}

	db, err := sql.Open("postgres", pool.dbConfig.ConnectionString())
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxConn; i++ {
		err := pool.openNewConnection(db)
		if err != nil {
			return nil, err
		}
	}
	return pool, nil
}

func (p *ConnectionPool) openNewConnection(db *sql.DB) error {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	p.conns <- conn
	return nil
}

func (p *ConnectionPool) Get() *sql.Conn {
	return <-p.conns
}

func (p *ConnectionPool) Put(conn *sql.Conn) {
	p.conns <- conn
}

func (p *ConnectionPool) Close() {
	for i := 0; i < p.maxConn; i++ {
		conn := <-p.conns
		conn.Close()
	}
}
