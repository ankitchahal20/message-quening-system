package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ankit/project/message-quening-system/internal/config"
)

func New() (postgres, error) {
	cfg := config.GetConfig()
	connString := "host=" + cfg.Database.Host + " " + "dbname=" + cfg.Database.DBname + " " + "password=" +
		cfg.Database.Password + " " + "user=" + cfg.Database.User + " " + "port=" + fmt.Sprint(cfg.Database.Port)

	conn, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Unable to connect: %v\n", err))
		return postgres{}, err
	}

	log.Println("Connected to database")

	err = conn.Ping()
	if err != nil {
		log.Fatal("Cannot Ping the database")
		return postgres{}, err
	}
	log.Println("pinged database")

	return postgres{db: conn}, nil
}
