package account

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/mail"
	"github.com/shigde/sfu/pkg/authentication"
	"golang.org/x/exp/slog"
)

const activateAccountURLPath = "activateAccount"

var ErrInvalidCredentials = errors.New("invalid credentials")

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
	name := account.User

	actor, err := models.NewPersonActor(s.instanceUrl, name)
	if err != nil {
		return fmt.Errorf("creating actor: %w", err)
	}

	account.Actor = actor

	// transform username in domain-specific UserID
	account.User = instance.BuildUserId(name, s.instanceUrl)

	_, err = s.repo.Add(ctx, account)
	if err != nil {
		return fmt.Errorf("adding account: %w", err)
	}

	token := NewEmailVerificationToken(account)

	if err = s.repo.AddVerificationToken(ctx, token); err != nil {
		return fmt.Errorf("creating verify token: %w", err)
	}

	link := s.instanceUrl.String() + "/" + activateAccountURLPath + "/" + token.UUID
	if err = s.mailSender.SendActivateAccountMail(account.User, account.Email, link); err != nil {
		return fmt.Errorf("sending verify email: %w", err)
	}

	return nil
}

func (s *AccountService) VerifyAccount(ctx context.Context, token string) error {
	if err := s.repo.RedeemAccountVerificationToken(ctx, token); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			slog.Warn("verifying account", "error", err)
			return err
		}
		slog.Error("verifying account", "error", err)
		return err
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

func (s *AccountService) GetAuthTokenByLogin(ctx context.Context, login *authentication.Login) (*authentication.Token, error) {
	account, err := s.repo.findByEmail(ctx, login.Email)
	if err != nil {
		return nil, fmt.Errorf("find login account: %w", err)
	}

	if valid := VerifyPassword(login.Pass, account.Password); !valid {
		return nil, ErrInvalidCredentials
	}

	token, err := session.CreateJWTToken(account.UUID, s.config.JWT)
	if err != nil {
		return nil, fmt.Errorf("create jwt token: %w", err)
	}

	return &authentication.Token{JWT: token}, nil
}

func (s *AccountService) GetAccount(ctx context.Context, userUuid *uuid.UUID) (*Account, error) {
	account, err := s.repo.findByUuid(ctx, userUuid)
	if err != nil {
		return nil, fmt.Errorf("find account by uuid: %w", err)
	}
	return account, nil
}
