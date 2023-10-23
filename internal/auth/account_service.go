package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/pkg/authentication"
)

type AccountService struct {
	config        *SecurityConfig
	instanceToken string
	repo          *accountRepository
}

func NewAccountService(store *storage.Store, instanceToken string, config *SecurityConfig) (*AccountService, error) {
	repo, err := newAccountRepository(store)
	if err != nil {
		return nil, fmt.Errorf("creating account repository: %w", err)
	}

	return &AccountService{
		config:        config,
		instanceToken: instanceToken,
		repo:          repo,
	}, nil
}

func (s *AccountService) CreateAccountByActor(actor *models.Actor) error {
	return nil
}

func (s *AccountService) DeleteAccountByActor(actor *models.Actor) error {
	return nil
}

func (s *AccountService) GetAuthToken(ctx context.Context, user *authentication.User) (*authentication.Token, error) {
	if user.Token != s.instanceToken {
		return nil, errors.New("invalid instance auth token")
	}
	account, err := s.repo.findByUserName(ctx, user.UserId)
	if err != nil {
		return nil, fmt.Errorf("find account: %w", err)
	}

	token, err := CreateJWTToken(account.UUID, s.config.JWT)
	if err != nil {
		return nil, fmt.Errorf("create jwt token: %w", err)
	}

	return &authentication.Token{JWT: token}, nil
}
