package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/whaleship/med-token/internal/dto"
	"github.com/whaleship/med-token/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type fakeRefreshRepo struct {
	sync.Mutex
	SaveFunc   func(ctx context.Context, jti, hash, ip, userID string, ttl time.Duration) error
	GetFunc    func(ctx context.Context, jti string) (hash, ip, userID string, err error)
	DeletedJTI string
}

func (f *fakeRefreshRepo) Save(ctx context.Context, jti, hash, ip, userID string, ttl time.Duration) error {
	if f.SaveFunc == nil {
		return nil
	}
	return f.SaveFunc(ctx, jti, hash, ip, userID, ttl)
}

func (f *fakeRefreshRepo) Get(ctx context.Context, jti string) (string, string, string, error) {
	if f.GetFunc == nil {
		return "", "", "", errors.New("not implemented")
	}
	return f.GetFunc(ctx, jti)
}

func (f *fakeRefreshRepo) Delete(ctx context.Context, jti string) error {
	f.Lock()
	defer f.Unlock()
	f.DeletedJTI = jti
	return nil
}

type fakeEmailService struct {
	sync.Mutex
	SentTo  []string
	Subject string
	Body    string
}

func (f *fakeEmailService) Send(to []string, subject, body string) error {
	f.Lock()
	defer f.Unlock()
	f.SentTo = to
	f.Subject = subject
	f.Body = body
	return nil
}

func TestCreateTokens(t *testing.T) {
	ctx := context.Background()
	secret := []byte("test-secret")
	userID := "user-123"
	ip := "1.2.3.4"

	var savedJTI, savedHash string
	fRepo := &fakeRefreshRepo{
		SaveFunc: func(ctx context.Context, jti, hash, ipAddr, uid string, ttl time.Duration) error {
			savedJTI, savedHash, _, _ = jti, hash, ipAddr, uid
			return nil
		},
	}
	emailSvc := &fakeEmailService{}
	svc := NewAuthService(secret, fRepo, emailSvc)

	t.Run("success", func(t *testing.T) {
		resp, err := svc.CreateTokens(ctx, userID, ip)
		if err != nil {
			t.Fatal(err)
		}
		if resp.AccessToken == "" {
			t.Error("access token is empty")
		}
		if resp.RefreshToken == "" {
			t.Error("refresh token is empty")
		}
		data, err := base64.StdEncoding.DecodeString(resp.RefreshToken)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.SplitN(string(data), ":", 2)
		if len(parts) != 2 {
			t.Error("refresh token format invalid")
		}
		raw := parts[1]
		if err := bcrypt.CompareHashAndPassword([]byte(savedHash), []byte(raw)); err != nil {
			t.Error("hash mismatch")
		}
		claims := &entity.Claims{}
		parsed, err := jwt.ParseWithClaims(resp.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !parsed.Valid {
			t.Fatal("invalid access token")
		}
		if claims.UID != userID {
			t.Error("unexpected UID")
		}
		if claims.IP != ip {
			t.Error("unexpected IP")
		}
		if claims.RefreshJTI != savedJTI {
			t.Error("unexpected RefreshJTI")
		}
	})

	t.Run("save error", func(t *testing.T) {
		repo := &fakeRefreshRepo{
			SaveFunc: func(ctx context.Context, jti, hash, ip, userID string, ttl time.Duration) error {
				return errors.New("fail")
			},
		}
		svc := NewAuthService(secret, repo, emailSvc)
		_, err := svc.CreateTokens(ctx, userID, ip)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestRefresh(t *testing.T) {
	ctx := context.Background()
	secret := []byte("test-secret")
	userID := "user-123"
	oldIP := "1.2.3.4"
	newIP := "5.6.7.8"

	rjti := "refresh-jti"
	raw := "raw-token-value"
	hash, _ := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	combined := fmt.Sprintf("%s:%s", rjti, raw)
	refreshToken := base64.StdEncoding.EncodeToString([]byte(combined))

	fRepo := &fakeRefreshRepo{
		GetFunc: func(ctx context.Context, jti string) (string, string, string, error) {
			return string(hash), oldIP, userID, nil
		},
		SaveFunc: func(ctx context.Context, jti, hash, ip, userID string, ttl time.Duration) error {
			return nil
		},
	}
	emailSvc := &fakeEmailService{}
	svc := NewAuthService(secret, fRepo, emailSvc)

	t.Run("invalid base64", func(t *testing.T) {
		_, err := svc.Refresh(ctx, dto.RefreshRequest{RefreshToken: "!!"}, newIP)
		if err == nil {
			t.Error("expected decode error")
		}
	})

	t.Run("bad structure", func(t *testing.T) {
		tok := base64.StdEncoding.EncodeToString([]byte("bad-token"))
		_, err := svc.Refresh(ctx, dto.RefreshRequest{RefreshToken: tok}, newIP)
		if err == nil {
			t.Error("expected structure error")
		}
	})

	t.Run("get error", func(t *testing.T) {
		repo := &fakeRefreshRepo{
			GetFunc: func(ctx context.Context, jti string) (string, string, string, error) {
				return "", "", "", errors.New("fail")
			},
		}
		svc := NewAuthService(secret, repo, emailSvc)
		_, err := svc.Refresh(ctx, dto.RefreshRequest{RefreshToken: refreshToken}, newIP)
		if err == nil {
			t.Error("expected get error")
		}
	})

	t.Run("invalid raw", func(t *testing.T) {
		bad := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", rjti, "bad")))
		_, err := svc.Refresh(ctx, dto.RefreshRequest{RefreshToken: bad}, newIP)
		if err == nil {
			t.Error("expected mismatch error")
		}
	})

	t.Run("ip changed sends email", func(t *testing.T) {
		resp, err := svc.Refresh(ctx, dto.RefreshRequest{RefreshToken: refreshToken}, newIP)
		if err != nil {
			t.Fatal(err)
		}
		emailSvc.Lock()
		defer emailSvc.Unlock()
		if len(emailSvc.SentTo) == 0 {
			t.Error("expected email")
		}
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Error("expected tokens")
		}
	})

	t.Run("success no ip change", func(t *testing.T) {
		emailSvc = &fakeEmailService{}
		svc := NewAuthService(secret, fRepo, emailSvc)
		resp, err := svc.Refresh(ctx, dto.RefreshRequest{RefreshToken: refreshToken}, oldIP)
		if err != nil {
			t.Fatal(err)
		}
		emailSvc.Lock()
		defer emailSvc.Unlock()
		if len(emailSvc.SentTo) != 0 {
			t.Error("unexpected email")
		}
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Error("expected tokens")
		}
		fRepo.Lock()
		defer fRepo.Unlock()
		if fRepo.DeletedJTI != rjti {
			t.Error("expected Delete call")
		}
	})
}
