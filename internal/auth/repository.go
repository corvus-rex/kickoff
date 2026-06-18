package auth

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

// FindUserByEmail looks up a non-deleted user by email (case-insensitive).
func FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	err := db.Where("email = ?", normalizeEmail(email)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}