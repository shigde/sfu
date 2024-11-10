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
const passwordForgetURLPath = "newPassword"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrAccountAlreadyExists = errors.New("account already exists")

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
	// transform username in domain-specific UserID
	account.User = instance.BuildUserId(name, s.instanceUrl)
	account.Role = USER

	// Check an account exists. If exists check for recreate or ignore
	// -------------------------------------------------------------------------------------------------------------------
	existAccount, err := s.repo.findByEmail(ctx, account.Email)
	if err != nil && !errors.Is(err, ErrAccountNotFound) {
		return fmt.Errorf("loocking for existing account: %w", err)
	}

	if existAccount != nil && existAccount.Active {
		// we ignore the try to recreate an active existing account
		return ErrAccountAlreadyExists
	}

	if existAccount != nil && !existAccount.Active {
		// we delete a previous created account if were not activated
		if err := s.repo.deleteByUuid(ctx, existAccount.UUID); err != nil {
			return fmt.Errorf("deleting existing inactive account: %w", err)
		}
	}

	// Create a new account.
	// -------------------------------------------------------------------------------------------------------------------
	actor, err := models.NewPersonActor(s.instanceUrl, name)
	if err != nil {
		return fmt.Errorf("creating actor: %w", err)
	}

	account.Actor = actor
	_, err = s.repo.Add(ctx, account)
	if err != nil {
		return fmt.Errorf("adding account: %w", err)
	}

	if err = s.sendVerificationMail(ctx, account); err != nil {
		return fmt.Errorf("verify new account: %w", err)
	}

	return nil
}

func (s *AccountService) sendVerificationMail(ctx context.Context, account *Account) error {
	token := NewEmailVerificationToken(account)

	if err := s.repo.AddVerificationToken(ctx, token); err != nil {
		return fmt.Errorf("creating verify token: %w", err)
	}

	link := s.instanceUrl.String() + "/" + activateAccountURLPath + "/" + token.UUID

	// Do not send an email if placeholder mail is used
	if IsPlaceholderEmail(account.Email) {
		return nil
	}

	if err := s.mailSender.SendActivateAccountMail(account.User, account.Email, link); err != nil {
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

func (s *AccountService) CreateForgotPasswordToken(ctx context.Context, email string) error {
	account, err := s.repo.findActiveByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrAccountNotFound) {
		return fmt.Errorf("serching for pass forgot accoun: %w", err)
	}

	token := NewPasswordVerificationToken(account)

	if err := s.repo.AddVerificationToken(ctx, token); err != nil {
		return fmt.Errorf("creating verify token for new password: %w", err)
	}

	link := s.instanceUrl.String() + "/" + passwordForgetURLPath + "/" + token.UUID

	if err := s.mailSender.SendActivateAccountMail(account.User, account.Email, link); err != nil {
		return fmt.Errorf("sending verify email for new password: %w", err)
	}

	return nil

}

func (s *AccountService) CreateAccountByActor(ctx context.Context, actor *models.Actor) error {
	return nil
}

func (s *AccountService) DeleteAccountByActor(ctx context.Context, actor *models.Actor) error {
	return nil
}

func (s *AccountService) GetAuthToken(ctx context.Context, user *authentication.ClientUser) (*authentication.Token, error) {
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
	account, err := s.repo.findActiveByEmail(ctx, login.Email)
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
	account, err := s.repo.findActiveByUuid(ctx, userUuid)
	if err != nil {
		return nil, fmt.Errorf("find account by uuid: %w", err)
	}
	return account, nil
}

func (s *AccountService) UpdatePassword(ctx context.Context, userUuid *uuid.UUID, oldPass string, newPass string) error {
	account, err := s.GetAccount(ctx, userUuid)
	if err != nil {
		return fmt.Errorf("find account by uuid for update pass: %w", err)
	}

	if valid := VerifyPassword(oldPass, account.Password); !valid {
		return ErrInvalidCredentials
	}

	hash, err := HashPassword(newPass)
	if err != nil {
		return fmt.Errorf("creating hash for update pass by uuid: %w", err)
	}
	account.Password = hash

	if err := s.repo.update(ctx, account); err != nil {
		return fmt.Errorf("updating account for new pass by uuid: %w", err)
	}

	return nil
}

func (s *AccountService) UpdatePasswordByToken(ctx context.Context, token string, newPass string) error {
	passToken, err := s.repo.RedeemPassForgetToken(ctx, token)
	if errors.Is(err, ErrTokenNotFound) {
		slog.Warn("redeem new pass token not valid", "error", err)
		return err
	}
	if err != nil {
		slog.Error("verifying redeem new pass token", "error", err)
		return err
	}

	account := passToken.Account

	hash, err := HashPassword(newPass)
	if err != nil {
		return fmt.Errorf("creating hash for update pass by token: %w", err)
	}
	account.Password = hash

	if err := s.repo.update(ctx, account); err != nil {
		return fmt.Errorf("updating account for new pass by token: %w", err)
	}

	return nil
}

func (s *AccountService) GetConfig() *session.SecurityConfig {
	return s.config
}

func (s *AccountService) DeleteAccount(ctx context.Context, userUuid *uuid.UUID) error {
	if err := s.repo.deleteByUuid(ctx, userUuid.String()); err != nil {
		return fmt.Errorf("delete account by uuid: %w", err)
	}
	return nil
}
