package google

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotConfigured      = errors.New("google oauth is not configured")
	ErrNotAuthorized      = errors.New("google account is not authorized")
	ErrInvalidSpreadsheet = errors.New("invalid google spreadsheet")
	ErrPermissionDenied   = errors.New("google permission denied")
	ErrRateLimit          = errors.New("google api rate limit exceeded")
	ErrAuthExpired        = errors.New("google access token is expired or invalid")
	ErrNetwork            = errors.New("google api network failure")
	ErrAPIStatus          = errors.New("google api returned an error")
	ErrInvalidState       = errors.New("invalid or expired oauth state")
	ErrCodeExchange       = errors.New("google oauth token exchange failed")
)

const (
	tokenEndpoint  = "https://oauth2.googleapis.com/token"
	authEndpoint   = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultScopes  = "https://www.googleapis.com/auth/spreadsheets.readonly"
	stateTTL       = 10 * time.Minute
	accountKey     = "account"
	stateKeyPrefix = "state:"
	accessRefresh  = 5 * time.Minute // refresh early
)

type Config struct {
	ClientID        string
	ClientSecret    string
	ProjectID       string
	RedirectURL     string
	RefreshToken    string
	TokenKey        string // encryption key for tokens at rest
	SuccessRedirect string
}

func ConfigFromEnv() Config {
	return Config{
		ClientID:        os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:    os.Getenv("GOOGLE_CLIENT_SECRET"),
		ProjectID:       os.Getenv("GOOGLE_PROJECT_ID"),
		RedirectURL:     os.Getenv("GOOGLE_REDIRECT_URL"),
		RefreshToken:    os.Getenv("GOOGLE_REFRESH_TOKEN"),
		TokenKey:        os.Getenv("GOOGLE_TOKEN_ENCRYPTION_KEY"),
		SuccessRedirect: os.Getenv("GOOGLE_OAUTH_SUCCESS_REDIRECT"),
	}
}

type TokenStore interface {
	Get(key string) (string, error) // "" when missing
	Set(key, data string, expiresAt *time.Time) error
	Delete(key string) error
}

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type TokenManager struct {
	cfg   Config
	store TokenStore
	http  *http.Client

	mu    sync.Mutex
	cache *OAuthToken
}

func NewTokenManager(cfg Config, store TokenStore) *TokenManager {
	return &TokenManager{
		cfg:   cfg,
		store: store,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *TokenManager) Configured() bool {
	return m.cfg.ClientID != "" && m.cfg.ClientSecret != ""
}

func (m *TokenManager) BeginAuth() (authURL, state string, err error) {
	if !m.Configured() {
		return "", "", ErrNotConfigured
	}
	if m.cfg.RedirectURL == "" {
		return "", "", ErrNotConfigured
	}
	state, err = randomHex(16)
	if err != nil {
		return "", "", err
	}
	exp := time.Now().Add(stateTTL)
	if err := m.store.Set(stateKeyPrefix+state, "pending", &exp); err != nil {
		return "", "", err
	}
	return m.AuthURL(state), state, nil
}

func (m *TokenManager) AuthURL(state string) string {
	uv := url.Values{}
	uv.Set("client_id", m.cfg.ClientID)
	uv.Set("redirect_uri", m.cfg.RedirectURL)
	uv.Set("response_type", "code")
	uv.Set("scope", defaultScopes)
	uv.Set("access_type", "offline")
	uv.Set("prompt", "consent")
	uv.Set("state", state)
	return authEndpoint + "?" + uv.Encode()
}

func (m *TokenManager) Exchange(code, state string) error {
	if state == "" {
		return ErrInvalidState
	}
	key := stateKeyPrefix + state
	raw, err := m.store.Get(key)
	if err != nil {
		return err
	}
	if raw == "" {
		return ErrInvalidState
	}
	_ = m.store.Delete(key)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", m.cfg.ClientID)
	form.Set("client_secret", m.cfg.ClientSecret)
	form.Set("redirect_uri", m.cfg.RedirectURL)

	tok, err := m.tokenRequest(form)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCodeExchange, err)
	}
	if tok.RefreshToken == "" {
		return fmt.Errorf("%w: no refresh_token returned", ErrCodeExchange)
	}
	return m.storeToken(accountKey, tok)
}

func (m *TokenManager) AccessToken() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache != nil && m.cache.AccessToken != "" && time.Now().Before(m.cache.ExpiresAt.Add(-accessRefresh)) {
		return m.cache.AccessToken, nil
	}

	refresh := m.cfg.RefreshToken
	if raw, _ := m.store.Get(accountKey); raw != "" {
		if tok, err := m.decryptToken(raw); err == nil {
			if tok.RefreshToken != "" {
				refresh = tok.RefreshToken
			}
			if tok.AccessToken != "" && time.Now().Before(tok.ExpiresAt.Add(-accessRefresh)) {
				m.cache = tok
				return tok.AccessToken, nil
			}
		}
	}

	if refresh == "" {
		return "", ErrNotAuthorized
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", m.cfg.ClientID)
	form.Set("client_secret", m.cfg.ClientSecret)
	form.Set("refresh_token", refresh)

	tok, err := m.tokenRequest(form)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNotAuthorized, err)
	}
	tok.RefreshToken = refresh
	_ = m.storeToken(accountKey, tok)
	m.cache = tok
	return tok.AccessToken, nil
}

func (m *TokenManager) tokenRequest(form url.Values) (*OAuthToken, error) {
	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := m.http.Do(req)
	if err != nil {
		return nil, ErrNetwork
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint status %d", resp.StatusCode)
	}
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.AccessToken == "" {
		return nil, errors.New("empty access token")
	}
	return &OAuthToken{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		TokenType:    out.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(out.ExpiresIn) * time.Second),
	}, nil
}

func (m *TokenManager) storeToken(key string, tok *OAuthToken) error {
	raw, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	ct, err := m.Encrypt(raw)
	if err != nil {
		return err
	}
	return m.store.Set(key, ct, nil)
}

func (m *TokenManager) decryptToken(raw string) (*OAuthToken, error) {
	plain, err := m.Decrypt(raw)
	if err != nil {
		return nil, err
	}
	var t OAuthToken
	if err := json.Unmarshal(plain, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (m *TokenManager) Encrypt(plain []byte) (string, error) {
	sum := sha256.Sum256([]byte(m.cfg.TokenKey))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, plain, nil)
	return "gcm1:" + base64.StdEncoding.EncodeToString(append(nonce, sealed...)), nil
}

func (m *TokenManager) Decrypt(ct string) ([]byte, error) {
	if !strings.HasPrefix(ct, "gcm1:") {
		return nil, errors.New("unrecognized ciphertext format")
	}
	sum := sha256.Sum256([]byte(m.cfg.TokenKey))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ct, "gcm1:"))
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, sealed := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	return gcm.Open(nil, nonce, sealed, nil)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
