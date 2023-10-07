package auth

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/storage"
)

type AccountService struct {
	repo *accountRepository
}

func NewAccountService(store *storage.Store) (*AccountService, error) {
	repo, err := newAccountRepository(store)
	if err != nil {
		return nil, fmt.Errorf("creating account repository: %w", err)
	}

	return &AccountService{
		repo: repo,
	}, nil
}

func (s *AccountService) CreateAccountByActor(actor *models.Actor) error {
	return nil
}

func (s *AccountService) DeleteAccountByActor(actor *models.Actor) error {
	return nil
}

func (s *AccountService) GetAuthToken(uuid *uuid.UUID, token string) (string, error) {
	return "", nil
}
