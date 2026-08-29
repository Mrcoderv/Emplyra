package services

import (
	"errors"
	"time"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/config"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrTokenInvalid       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	users  *repositories.UserRepository
	tokens *repositories.TokenRepository
	jwt    *auth.JWTManager
	cfg    *config.Config
	audit  *auditmanager.Service
}

func NewAuthService(users *repositories.UserRepository, tokens *repositories.TokenRepository, jwt *auth.JWTManager, cfg *config.Config, audit *auditmanager.Service) *AuthService {
	return &AuthService{users: users, tokens: tokens, jwt: jwt, cfg: cfg, audit: audit}
}

type LoginInput struct {
	Identifier string
	Password   string
	IP         string
	UserAgent  string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func (s *AuthService) Login(in LoginInput) (*models.User, *TokenPair, error) {
	user, err := s.users.FindByEmailOrUsername(in.Identifier)
	if err != nil {
		s.audit.FailedLogin(in.Identifier, in.IP, in.UserAgent)
		return nil, nil, ErrInvalidCredentials
	}
	if !auth.VerifyPassword(user.PasswordHash, in.Password) {
		s.audit.FailedLogin(in.Identifier, in.IP, in.UserAgent)
		return nil, nil, ErrInvalidCredentials
	}
	if !auth.UserStatusAllowed(user.Status) {
		s.audit.Record(user.ID, models.ActionLogin, "user", user.ID, in.IP, in.UserAgent, map[string]string{"reason": "account_disabled"})
		return nil, nil, ErrAccountDisabled
	}

	pair, err := s.issueTokens(user, in.IP, in.UserAgent)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	_ = s.users.Update(user.ID, map[string]interface{}{"last_login_at": &now})
	s.audit.Record(user.ID, models.ActionLogin, "user", user.ID, in.IP, in.UserAgent, nil)
	return user, pair, nil
}

func (s *AuthService) Refresh(refreshToken, ip, userAgent string) (*models.User, *TokenPair, error) {
	hash := auth.HashRefreshToken(refreshToken)
	stored, err := s.tokens.FindByHash(hash)
	if err != nil || stored.UserID == "" {
		return nil, nil, ErrTokenInvalid
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return nil, nil, ErrTokenInvalid
	}
	user, err := s.users.FindByID(stored.UserID)
	if err != nil || !auth.UserStatusAllowed(user.Status) {
		return nil, nil, ErrAccountDisabled
	}
	pair, err := s.issueTokens(user, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}
	_ = s.tokens.Revoke(stored.ID, "")
	s.audit.Record(user.ID, models.ActionLogin, "user", user.ID, ip, userAgent, map[string]string{"action": "refresh"})
	return user, pair, nil
}

func (s *AuthService) Logout(refreshToken, ip, userAgent string) error {
	hash := auth.HashRefreshToken(refreshToken)
	stored, err := s.tokens.FindByHash(hash)
	if err != nil {
		return nil
	}
	_ = s.tokens.Revoke(stored.ID, "")
	if stored.UserID != "" {
		s.audit.Record(stored.UserID, models.ActionLogout, "user", stored.UserID, ip, userAgent, nil)
	}
	return nil
}

func (s *AuthService) Me(userID string) (*models.User, []string, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}
	perms, err := s.Users().RolePermissions(user.RoleID)
	if err != nil {
		return nil, nil, err
	}
	return user, perms, nil
}

func (s *AuthService) issueTokens(user *models.User, ip, userAgent string) (*TokenPair, error) {
	access, err := s.jwt.Generate(user.ID, user.Username, user.Role.Name, user.RoleID)
	if err != nil {
		return nil, err
	}
	raw := auth.GenerateRefreshToken()
	now := time.Now()
	t := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: auth.HashRefreshToken(raw),
		ExpiresAt: now.Add(s.cfg.RefreshTokenTTL),
		IP:        ip,
		UserAgent: userAgent,
	}
	if err := s.tokens.Create(t); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) Users() *repositories.UserRepository { return s.users }
func (s *AuthService) JWT() *auth.JWTManager               { return s.jwt }
