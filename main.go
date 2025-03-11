package main

import (
	"log"
	"strings"

	"github.com/drawref/drawref-backend/drawref"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// loading config
	config, err := drawref.LoadConfig()
	if err != nil {
		log.Fatal("LoadConfig failed:", err)
	}

	// setup s3
	err = drawref.OpenS3(config.UploadS3Bucket, config.UploadS3KeyPrefix, config.UploadS3URLPrefix)
	if err != nil {
		log.Fatal("OpenS3 failed:", err)
	}

	// upgrading db
	m, err := migrate.New("file://"+config.DatabaseMigrationsPath, config.DatabaseUrl)
	if err != nil {
		log.Fatal("DB Migrations loading failed:", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("DB Migration failure:", err)
	}
	m.Close()

	// open db
	err = drawref.OpenDatabase(config.DatabaseUrl)
	if err != nil {
		log.Fatal("Opening db failed:", err)
	}

	// api router
	router := drawref.GetRouter(nil, strings.Split(config.CorsAllowedFrom, " "))
	router.Run(config.Address)
}
