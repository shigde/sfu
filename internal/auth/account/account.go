package account

import (
	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"gorm.io/gorm"
)

type Account struct {
	// User: username@domian  equal to `${Actor.PreferredUsername}@domain`
	User     string        `json:"user"     gorm:"index;unique"`
	Email    string        `json:"email"    gorm:"index;unique"`
	UUID     string        `json:"-"        gorm:"index;unique"`
	Role     Role          `json:"-"        gorm:"not null,default:2"`
	Password string        `json:"password" gorm:"not null"`
	Active   bool          `json:"-"        gorm:"not null,default:false"`
	ActorId  uint          `json:"-"        gorm:"not null;unique"`
	Actor    *models.Actor `json:"-"        gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}

type Role int32

const (
	ADMIN Role = iota + 1
	USER
	GUEST
	SERVICE
)

type VerificationToken struct {
	AccountId uint      `gorm:"not null"`
	Account   *Account  `gorm:"foreignKey:AccountId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Type      TokenType `gorm:"not null"`
	UUID      string    `gorm:"index;unique"`
	Verified  bool      `gorm:"not null,default:false"`
	gorm.Model
}

type TokenType int32

const (
	EMAIL    TokenType = 0
	PASSWORD TokenType = 1
)

func NewEmailVerificationToken(account *Account) *VerificationToken {
	return &VerificationToken{
		Account: account,
		UUID:    uuid.NewString(),
		Type:    EMAIL,
	}
}

func NewPasswordVerificationToken(account *Account) *VerificationToken {
	return &VerificationToken{
		Account: account,
		UUID:    uuid.NewString(),
		Type:    PASSWORD,
	}
}

func RoleToString(role Role) string {
	switch role {
	case USER:
		return "user"
	case GUEST:
		return "guest"
	case ADMIN:
		return "admin"
	case SERVICE:
		return "service"
	default:
		return "guest"
	}
}
