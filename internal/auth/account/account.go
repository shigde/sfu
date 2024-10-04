package account

import (
	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"gorm.io/gorm"
)

type Account struct {
	User     string        `json:"user"     gorm:"index;unique"`
	Email    string        `json:"email"    gorm:"index;unique"`
	UUID     string        `json:"-"        gorm:"index;unique"`
	Type     AccountType   `json:"-"        gorm:"not null,default:2"`
	Password string        `json:"password" gorm:"not null"`
	Active   bool          `json:"-"        gorm:"not null,default:false"`
	ActorId  uint          `json:"-"        gorm:"not null;unique"`
	Actor    *models.Actor `json:"-"        gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}

type AccountType int32

const (
	ADMIN AccountType = 1
	USER  AccountType = 2
	GUEST AccountType = 3
)

type AccountVerificationToken struct {
	AccountId uint      `gorm:"not null"`
	Account   *Account  `gorm:"foreignKey:AccountId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Type      TokenType `gorm:"not null"`
	UUID      string    `gorm:"index;unique"`
	gorm.Model
}

type TokenType int32

const (
	EMAIL    TokenType = 0
	PASSWORD TokenType = 1
)

func NewEmailVerificationToken(account *Account) *AccountVerificationToken {
	return &AccountVerificationToken{
		Account: account,
		UUID:    uuid.NewString(),
		Type:    EMAIL,
	}
}

func NewPasswordVerificationToken(account *Account) *AccountVerificationToken {
	return &AccountVerificationToken{
		Account: account,
		UUID:    uuid.NewString(),
		Type:    PASSWORD,
	}
}
