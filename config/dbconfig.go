package config

import (
	"fmt"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DBConnection() (*gorm.DB, error) {

	dbURL := "postgres://avnadmin:AVNS_CACPBumpuKG1HivSydo@pg-4de421b-taskqueue.k.aivencloud.com:17326/defaultdb?sslmode=require"
	fmt.Println("url is", dbURL)
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL is not set")
	}

	db, err := gorm.Open(
		postgres.Open(dbURL),
		&gorm.Config{},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"database connection failed: %w",
			err,
		)
	}

	fmt.Println("PostgreSQL connected successfully")

	return db, nil

}
