package main

import (
	"context"
	"fmt"
	"modules"
)

func Run() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPostgreConfig()

	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	fmt.Println(_postgre)
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.POstgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "postgres",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}
