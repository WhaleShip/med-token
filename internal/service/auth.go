package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/whaleship/med-token/internal/dto"
	"github.com/whaleship/med-token/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type EmailService interface {
	Send(to []string, subject, body string) error
}

type RefreshRepo interface {
	Save(ctx context.Context, jti, hash, ip, userID string, ttl time.Duration) error
	Get(ctx context.Context, jti string) (hash, ip, userID string, err error)
	Delete(ctx context.Context, jti string) error
}

type authService struct {
	secret       []byte
	refreshR     RefreshRepo
	emailService EmailService
}

func NewAuthService(secret []byte, repo RefreshRepo, es EmailService) *authService {
	return &authService{secret, repo, es}
}

func (s *authService) CreateTokens(ctx context.Context, userID, ip string) (*dto.TokenResponse, error) {
	jti := uuid.NewString()

	rawSecret := uuid.NewString()
	combined := fmt.Sprintf("%s:%s", jti, rawSecret)
	refreshEncoded := base64.StdEncoding.EncodeToString([]byte(combined))

	hashed, err := bcrypt.GenerateFromPassword([]byte(rawSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	if err := s.refreshR.Save(ctx, jti, string(hashed), ip, userID, 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("refreshR.Save: %w", err)
	}

	now := time.Now()
	claims := &entity.Claims{
		UID:        userID,
		IP:         ip,
		RefreshJTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessSigned, err := token.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("SignedString: %w", err)
	}

	return &dto.TokenResponse{
		AccessToken:  accessSigned,
		RefreshToken: refreshEncoded,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshRequest, ip string) (*dto.TokenResponse, error) {
	data, err := base64.StdEncoding.DecodeString(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("bad structure")
	}
	jti, raw := parts[0], parts[1]

	hash, oldIP, userID, err := s.refreshR.Get(ctx, jti)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)); err != nil {
		return nil, err
	}
	if ip != oldIP {
		go func() {
			err := s.emailService.Send(
				[]string{"andreipogirei@yandex.ru"},
				"Warning: IP address changed",
				fmt.Sprintf("Your session IP changed from %s to %s", oldIP, ip),
			)
			if err != nil {
				log.Println(err)
			}
		}()
	}
	if err = s.refreshR.Delete(ctx, jti); err != nil {
		return nil, err
	}

	return s.CreateTokens(ctx, userID, ip)
}
