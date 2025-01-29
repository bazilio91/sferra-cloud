package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

func seed() error {
	if err := initAdminUser(); err != nil {
		return err
	}

	if err := seedClient(); err != nil {
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

func seedClient() error {
	// check if client id = 1 exists
	var client proto.ClientORM
	err := db.DB.First(&client, 1).Error
	if err == nil {
		return nil
	}
	// create a default client
	client = proto.ClientORM{
		ClientUserId: nil,
		Id:           1,
		Inn:          "123123123123123123",
		Name:         "Demacia",
		Ogrn:         "3213213213213213213",
		OwnerFio:     "Jarvan IV Lightshield",
		Quota:        100,
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
	}).Error
	if err != nil {
		return err
	}

	log.Printf("User for default client created\n")

	return nil
}
