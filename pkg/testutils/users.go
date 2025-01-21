// pkg/testutils/data.go
package testutils

import (
	"fmt"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

func CreateTestClient(DB *gorm.DB, name string, quota int64) (*proto.Client, error) {
	client := &proto.Client{
		Name:  name,
		Quota: quota,
	}
	if err := DB.Create(client).Error; err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return client, nil
}

func CreateTestUser(DB *gorm.DB, email, password string, clientID uint64) (*proto.ClientUser, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := &proto.ClientUser{
		Email:    email,
		Password: string(hashedPassword),
		ClientId: clientID,
	}
	if err := DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return user, nil
}
