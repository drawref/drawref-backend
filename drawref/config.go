package drawref

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	GithubClientID         string `env:"GITHUB_CLIENT_ID"`
	GithubClientSecret     string `env:"GITHUB_CLIENT_SECRET"`
	GithubAdminUIDsRaw     string `env:"GITHUB_ADMIN_UIDS"`
	GithubAdminUIDs        []string
	AllowLocalImagesRaw    string `env:"ALLOW_LOCAL_IMAGES"`
	AllowLocalImages       bool
	LocalImagesRootPath    string `env:"LOCAL_IMAGES_ROOT_PATH"`
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

	config.AllowLocalImages, err = strconv.ParseBool(config.AllowLocalImagesRaw)
	if err != nil || config.LocalImagesRootPath == "" {
		config.AllowLocalImages = false
	}
	if config.AllowLocalImages {
		fmt.Println("Allowing local images from", config.LocalImagesRootPath)
	}

	for _, uid := range strings.Split(config.GithubAdminUIDsRaw, " ") {
		if uid != "" {
			config.GithubAdminUIDs = append(config.GithubAdminUIDs, uid)
		}
	}
	fmt.Println("Allowing GitHub UIDs as admin: [", strings.Join(config.GithubAdminUIDs, ", "), "]")

	return config, nil
}
