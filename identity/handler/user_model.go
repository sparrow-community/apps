package handler

import (
	"github.com/sparrow-community/app/identity/config"
	"github.com/sparrow-community/protos/identity"
)

// validate

const passwordRule = "required,password,min=8,max=16"

var UsernameSignUpRequestRules = map[string]string{
	"Identifier": "required",
	"Password":   passwordRule,
	"Type":       "required,eq=1",
}

var EmailSignUpRequestRules = map[string]string{
	"Identifier": "required,email",
	"Password":   passwordRule,
	"Type":       "required,eq=2",
}

var PhoneSignUpRequestRules = map[string]string{
	"Identifier": "required,e164",
	"Password":   passwordRule,
	"Type":       "required,eq=3",
}

var SignUpRequestRules = map[string]string{
	"Identifier": "required",
	"Password":   passwordRule,
	"Type":       "required,min=1,max=3",
}

func getSignUpRule(t identity.UserCredentialType) map[string]string {
	rules := SignUpRequestRules
	switch t {
	case identity.UserCredentialType_USERNAME:
		rules = UsernameSignUpRequestRules
	case identity.UserCredentialType_EMAIL:
		rules = EmailSignUpRequestRules
	case identity.UserCredentialType_PHONE:
		rules = PhoneSignUpRequestRules
	}
	return rules
}

// User .
type User struct {
	ID         string `json:"id" gorm:"primaryKey"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	DisabledAt int64  `json:"disabledAt"`
	DeletedAt  int64  `json:"deletedAt"`
	CreatedAt  int64  `json:"createdAt"`
}

// UserCredential .
type UserCredential struct {
	ID             int64                       `json:"id" gorm:"primaryKey,autoIncrement"`
	UserID         string                      `json:"UserID"`
	Type           identity.UserCredentialType `json:"type"`
	Identifier     string                      `json:"identifier"`
	Salt           []byte                      `json:"salt"`
	SecretData     string                      `json:"secretData"`
	Verified       bool                        `json:"verified"`
	VerificationAt int64                       `json:"verificationAt"`
	CreatedAt      int64                       `json:"createdAt"`
}

func (uc *UserCredential) IdentifierExists() (bool, error) {
	var exists bool
	err := config.Conf.DB.Model(&UserCredential{}).Select("count(*) > 0").Where("identifier = ?", uc.Identifier).Find(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (uc *UserCredential) FindByIdentifier() error {
	return config.Conf.DB.Model(&UserCredential{}).Where("identifier = ?", uc.Identifier).Scan(&uc).Error
}
