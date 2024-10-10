package account

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/mail"
	"github.com/shigde/sfu/pkg/authentication"
	"golang.org/x/exp/slog"
)

type AccountService struct {
	config        *session.SecurityConfig
	instanceToken string
	instanceUrl   *url.URL
	mailSender    *mail.SenderService
	repo          *AccountRepository
}

func NewAccountService(
	repo *AccountRepository,
	instanceToken string,
	instanceUrl *url.URL,
	config *session.SecurityConfig,
	mail *mail.SenderService,
) *AccountService {
	return &AccountService{
		config:        config,
		instanceToken: instanceToken,
		instanceUrl:   instanceUrl,
		mailSender:    mail,
		repo:          repo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, account *Account) error {

	actor, err := models.NewPersonActor(s.instanceUrl, account.User)
	if err != nil {
		return fmt.Errorf("creating actor: %w", err)
	}

	account.Actor = actor
	_, err = s.repo.Add(ctx, account)
	if err != nil {
		return fmt.Errorf("adding account: %w", err)
	}

	token := NewEmailVerificationToken(account)

	if err = s.repo.AddVerificationToken(ctx, token); err != nil {
		return fmt.Errorf("creating verify token: %w", err)
	}

	if err = s.mailSender.SendActivateAccountMail(account.User, account.Email, token.UUID); err != nil {
		return fmt.Errorf("sending verify email: %w", err)
	}

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

	token, err := session.CreateJWTToken(account.UUID, s.config.JWT)
	if err != nil {
		return nil, fmt.Errorf("create jwt token: %w", err)
	}

	return &authentication.Token{JWT: token}, nil
}