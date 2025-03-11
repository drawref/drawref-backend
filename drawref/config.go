package drawref

import (
	"context"
	"errors"
	"fmt"
	"os"

	"aidanwoods.dev/go-paseto"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Address                string `env:"ADDR,required"`
	CorsAllowedFrom        string `env:"CORS_ALLOWED_FROM,required"`
	DatabaseUrl            string `env:"DATABASE_URL,required"`
	DatabaseMigrationsPath string `env:"DATABASE_MIGRATIONS_PATH,default=migrations"`
	PasetoKey              string `env:"PASETO_KEY"`
	UploadS3Bucket         string `env:"UPLOAD_S3_BUCKET, required"`
	UploadS3KeyPrefix      string `env:"UPLOAD_KEY_PREFIX, required"`
	UploadS3URLPrefix      string `env:"UPLOAD_URL_PREFIX, required"`
}

func LoadConfig() (config Config, err error) {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			return config, err
		}
	}

	ctx := context.Background()

	if err := envconfig.Process(ctx, &config); err != nil {
		return config, err
	}

	if config.PasetoKey == "" {
		fmt.Println("New random PASETO key:", paseto.NewV4SymmetricKey().ExportHex())
		return config, errors.New("PASETO_KEY environment variable must be set")
	}

	return config, nil
}
