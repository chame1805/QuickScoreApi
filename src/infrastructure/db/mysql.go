package db

import (
"database/sql"
"fmt"
"os"

_ "github.com/go-sql-driver/mysql"
)

func Connect() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
getEnv("DB_USER", "apiuser"),
getEnv("DB_PASSWORD", "apipassword"),
getEnv("DB_HOST", "localhost"),
getEnv("DB_PORT", "3306"),
getEnv("DB_NAME", "apidb"),
)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir conexión: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar con MySQL: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)

	return conn, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
