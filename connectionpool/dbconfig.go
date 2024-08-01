package main

import (
	"fmt"
)

type DbConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func (c *DbConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database)
}

func NewDbConfig() *DbConfig {
	return &DbConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "postgres",
		Database: "postgres",
	}
}
