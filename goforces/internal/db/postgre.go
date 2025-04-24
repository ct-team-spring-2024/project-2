package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"time"
)

func ConnectToDB() {
	connStr := "postgres://postgres:example@localhost:5432/postgres"

	// Connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pgx.Connect(ctx, connStr)
	if err != nil {
		logrus.Error("Unable to connect to database: %v\n", err)
	}
	//defer conn.Close(ctx)

	logrus.Info("Connected to Postgre")

	// Sample query: Get current time
	//	var now time.Time
	// err = conn.QueryRow(ctx, "SELECT NOW()").Scan(&now)
	// if err != nil {
	// 	logrus.Error("Query failed: %v\n", err)
	// }

	// logrus.Info("Current time from DB: %v\n", now)
}
