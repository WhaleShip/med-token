package handlers

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/med-token/internal/dto"
)

type AuthService interface {
	CreateTokens(ctx context.Context, userID, ip string) (*dto.TokenResponse, error)
	Refresh(ctx context.Context, req dto.RefreshRequest, ip string) (*dto.TokenResponse, error)
}

type AuthHandler struct {
	svc AuthService
}

func NewAuthHandler(s AuthService) *AuthHandler {
	return &AuthHandler{svc: s}
}

func (h *AuthHandler) GetTokens(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "user_id required"})
	}
	resp, err := h.svc.CreateTokens(c.UserContext(), userID, c.IP())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	resp, err := h.svc.Refresh(c.UserContext(), req, c.IP())
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}
