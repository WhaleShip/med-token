package service

import (
	"context"
	"encoding/base64"
	"fmt"
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
	secret   []byte
	refreshR RefreshRepo
}

func NewAuthService(secret []byte, repo RefreshRepo) *authService {
	return &authService{secret, repo}
}

func (s *authService) CreateTokens(ctx context.Context, userID, ip string) (*dto.TokenResponse, error) {
	rjti := uuid.NewString()
	raw := uuid.NewString()
	combined := fmt.Sprintf("%s:%s", rjti, raw)
	encoded := base64.StdEncoding.EncodeToString([]byte(combined))

	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if err := s.refreshR.Save(ctx, rjti, string(hash), ip, userID, 7*24*time.Hour); err != nil {
		return nil, err
	}

	claims := &entity.Claims{
		UID:        userID,
		IP:         ip,
		RefreshJTI: rjti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	access, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  access,
		RefreshToken: encoded,
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

	hash, _, userID, err := s.refreshR.Get(ctx, jti)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)); err != nil {
		return nil, err
	}

	if err = s.refreshR.Delete(ctx, jti); err != nil {
		return nil, err
	}

	return s.CreateTokens(ctx, userID, ip)
}
