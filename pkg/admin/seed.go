package admin

import (
	"context"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/bazilio91/sferra-cloud/pkg/services/image"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"path"
)

func seed(s3Client *storage.S3Client) error {
	if err := initAdminUser(); err != nil {
		return err
	}

	if err := seedClient(s3Client); err != nil {
		return err
	}

	return nil
}

func initAdminUser() error {
	var count int64
	db.DB.Model(&proto.Admin{}).Count(&count)
	if count == 0 {
		// Create a default admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := proto.Admin{
			Username: "admin",
			Password: string(hashedPassword),
		}
		if err := db.DB.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("Default admin user created with username 'admin' and password '%s'\n", os.Getenv("ADMIN_PASSWORD"))
	}
	return nil
}

func seedClient(s3client *storage.S3Client) error {
	// check if client id = 1 exists
	var client proto.ClientORM
	err := db.DB.First(&client, 1).Error
	if err == nil {
		return nil
	}
	// create a default client
	client = proto.ClientORM{
		Id:         1,
		Inn:        "123123123123123123",
		Name:       "Demacia",
		Ogrn:       "3213213213213213213",
		OwnerFio:   "Jarvan IV Lightshield",
		TotalQuota: 1000,
		Quota:      100,
	}
	if err := db.DB.Create(&client).Error; err != nil {
		return err
	}

	log.Printf("Default client created with id 1\n")

	// add users for client
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = db.DB.Create(&proto.ClientUserORM{
		Client:   &client,
		Email:    "user@example.com",
		Password: string(hashedPassword),
		Username: "Test User",
		Role:     "Role",
	}).Error
	if err != nil {
		return err
	}

	log.Printf("User for default client created\n")

	task := &proto.DataRecognitionTaskORM{
		ClientId:                   &client.Id,
		FrontendResult:             nil,
		FrontendResultUnrecognized: nil,
		ProcessedImages:            nil,
		RecognitionResult:          nil,
		SourceImages:               nil,
		Status:                     0,
	}
	// create example recognition task
	err = db.DB.Create(&task).Error

	if err != nil {
		return err
	}

	// now lets upload some images
	imagePaths := []string{"0-4851.07  Љаоз®Є.jpeg", "016.72.281.jpeg"}
	imageService := image.NewService(s3client)

	for _, imagePath := range imagePaths {
		imageF, err := os.Open(path.Join("static", "test_data", imagePath))
		if err != nil {
			return err
		}

		taskImage, err := imageService.UploadTaskImage(context.Background(), client.Id, task.Id, imagePath, imageF)
		if err != nil {
			return err
		}

		task.SourceImages = append(task.SourceImages, taskImage.ID)
	}

	err = db.DB.Save(&task).Error
	if err != nil {
		return err
	}

	return nil
}
