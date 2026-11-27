package jwks

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
)

// KeySet は Authenticator の /.well-known/jwks.json から取得した公開鍵を保持する。
type KeySet struct {
	url    string
	client *http.Client
	mu     sync.RWMutex
	key    ed25519.PublicKey
}

// NewKeySet は起動時に公開鍵を取得する。失敗した場合は検証時に再取得する。
func NewKeySet() *KeySet {
	ks := &KeySet{
		url:    config.AUTHENTICATOR_URL + "/.well-known/jwks.json",
		client: &http.Client{Timeout: 5 * time.Second},
	}
	if err := ks.fetch(); err != nil {
		log.Printf("公開鍵の取得に失敗しました: %v", err)
	}
	return ks
}

// Parse は JWT を検証する。署名を検証できなかった場合は公開鍵を再取得して1度だけ再試行する。
func (ks *KeySet) Parse(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := ks.parse(tokenString, claims)
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) && !errors.Is(err, jwt.ErrTokenUnverifiable) {
		return token, err
	}
	if fetchErr := ks.fetch(); fetchErr != nil {
		return nil, fmt.Errorf("%w: %w", err, fetchErr)
	}
	return ks.parse(tokenString, claims)
}

func (ks *KeySet) parse(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(*jwt.Token) (any, error) {
		ks.mu.RLock()
		defer ks.mu.RUnlock()
		if ks.key == nil {
			return nil, domainerrors.ErrPublicKeyNotFetched
		}
		return ks.key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}))
}

func (ks *KeySet) fetch() error {
	res, err := ks.client.Get(ks.url)
	if err != nil {
		return fmt.Errorf("%w: %w", domainerrors.ErrFailedToFetchPublicKey, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", domainerrors.ErrFailedToFetchPublicKey, res.StatusCode)
	}

	var body struct {
		Keys []struct {
			Kty string `json:"kty"`
			Crv string `json:"crv"`
			X   string `json:"x"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return fmt.Errorf("%w: %w", domainerrors.ErrInvalidPublicKey, err)
	}
	if len(body.Keys) == 0 || body.Keys[0].Kty != "OKP" || body.Keys[0].Crv != "Ed25519" {
		return domainerrors.ErrInvalidPublicKey
	}
	key, err := base64.RawURLEncoding.DecodeString(body.Keys[0].X)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return domainerrors.ErrInvalidPublicKey
	}

	ks.mu.Lock()
	ks.key = key
	ks.mu.Unlock()
	return nil
}
