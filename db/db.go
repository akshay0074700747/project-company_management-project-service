package db

import (
	"fmt"
	"log"

	"github.com/akshay0074700747/project-company_management-project-service/config"
	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(cfg config.Config) *gorm.DB {

	psqlInfo := fmt.Sprintf("host=%s user=%s dbname=%s port=%s password=%s", cfg.DBhost, cfg.DBuser, cfg.DBname, cfg.DBport, cfg.DBpassword)
	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		log.Fatal("cannot connect to the db ", err)
	}

	db.AutoMigrate(&entities.Credentials{})
	db.AutoMigrate(&entities.Companies{})
	db.AutoMigrate(&entities.Members{})
	db.AutoMigrate(&entities.Owners{})
	db.AutoMigrate(&entities.MemberStatus{})
	db.AutoMigrate(&entities.TaskAssignations{})
	db.AutoMigrate(&entities.TaskStatuses{})
	db.AutoMigrate(&entities.NonTechnicalTaskDetials{})
	db.AutoMigrate(&entities.Issues{})
	db.AutoMigrate(&entities.Ratings{})
	db.AutoMigrate(&entities.Extensions{})

	return db
}

func ConnectMinio(cfg config.Config) *minio.Client {
	minioClient, err := minio.New(cfg.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return minioClient
}
