package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewDatabaseFromConfig creates a database object from configuration structure
func NewDatabaseFromConfig(config DatabaseConfig) (*gorm.DB, error) {
	switch config.Type {
	case "postgres":
		return gorm.Open(postgres.New(postgres.Config{DSN: config.DSN}), &gorm.Config{})
	case "mysql":
		return gorm.Open(mysql.New(mysql.Config{DSN: config.DSN}), &gorm.Config{})
	default:
		return gorm.Open(sqlite.Open(config.DSN), &gorm.Config{})
	}
}

// NewDatabaseFromFile creates a database object from path to file (passed as argument)
func NewDatabaseFromFile(path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}
