package handler

import (
	"net/http"
	"time"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/user"
	"labshop/internal/usecase/auth"
)

type AuthHandler struct {
	authUC *auth.Usecase
}

func NewAuthHandler(authUC *auth.Usecase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateMeRequest struct {
	Name    *string `json:"name,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Address *string `json:"address,omitempty"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toUserDTO(u *user.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Phone:     u.Phone,
		Address:   u.Address,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type AuthResponse struct {
	User  UserDTO `json:"user"`
	Token string  `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	u, token, err := h.authUC.Register(r.Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Phone:    req.Phone,
		Address:  req.Address,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, AuthResponse{
		User:  toUserDTO(u),
		Token: token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	u, token, err := h.authUC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, AuthResponse{
		User:  toUserDTO(u),
		Token: token,
	})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	u, err := h.authUC.GetMe(r.Context(), claims.UserID)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toUserDTO(u))
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	var req UpdateMeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	u, err := h.authUC.UpdateMe(r.Context(), claims.UserID, auth.UpdateMeInput{
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toUserDTO(u))
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	var req ChangePasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	if err := h.authUC.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}
