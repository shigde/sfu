package auth

import (
  "context"
  "errors"
  "fmt"

  "github.com/shigde/sfu/internal/activitypub/models"
  "github.com/shigde/sfu/internal/mail"
  "github.com/shigde/sfu/pkg/authentication"
  "golang.org/x/exp/slog"
)

type AccountService struct {
  config        *SecurityConfig
  instanceToken string
  mail          *mail.SenderService
  repo          *AccountRepository
}

func NewAccountService(repo *AccountRepository, instanceToken string, config *SecurityConfig) *AccountService {
  return &AccountService{
    config:        config,
    instanceToken: instanceToken,
    repo:          repo,
  }
}

func (s *AccountService) CreateAccount(ctx context.Context, account *Account) error {
  _, err := s.repo.Add(ctx, account)
  if err != nil {
    return fmt.Errorf("failed create account: %w", err)
  }
  s.mail.SendRegisterMail(account.User, account.Email, 'tokkkkkkkkkkkkken')
  return nil
}

func (s *AccountService) CreateAccountByActor(ctx context.Context, actor *models.Actor) error {
  return nil
}

func (s *AccountService) DeleteAccountByActor(ctx context.Context, actor *models.Actor) error {
  return nil
}

func (s *AccountService) GetAuthToken(ctx context.Context, user *authentication.User) (*authentication.Token, error) {
  slog.Debug("Auth", "Token", user.Token, "instance Token", s.instanceToken)
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
