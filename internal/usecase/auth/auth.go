package auth

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"labshop/internal/domain/errs"
	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
	Address  string
}

type UpdateMeInput struct {
	Name    *string
	Phone   *string
	Address *string
}

type Usecase struct {
	userRepo port.UserRepository
	hasher   port.PasswordHasher
	issuer   port.TokenIssuer
	idGen    port.IDGenerator
}

func New(userRepo port.UserRepository, hasher port.PasswordHasher, issuer port.TokenIssuer, idGen port.IDGenerator) *Usecase {
	return &Usecase{
		userRepo: userRepo,
		hasher:   hasher,
		issuer:   issuer,
		idGen:    idGen,
	}
}

func (u *Usecase) Register(ctx context.Context, in RegisterInput) (*user.User, string, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, "", errs.NewValidation("invalid email", errs.FieldViolation{Field: "email", Issue: "valid email required"})
	}

	if utf8.RuneCountInString(in.Password) < 10 {
		return nil, "", errs.NewValidation("password too short", errs.FieldViolation{Field: "password", Issue: "must be at least 10 characters"})
	}

	existing, err := u.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, "", errs.NewConflict("email already registered")
	}

	hash, err := u.hasher.Hash(in.Password)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	newUser := &user.User{
		ID:           u.idGen.NewID(),
		Email:        email,
		PasswordHash: hash,
		Name:         strings.TrimSpace(in.Name),
		Phone:        strings.TrimSpace(in.Phone),
		Address:      strings.TrimSpace(in.Address),
		Role:         user.RoleCustomer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := newUser.Validate(); err != nil {
		return nil, "", err
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, "", err
	}

	token, err := u.issuer.IssueToken(newUser)
	if err != nil {
		return nil, "", err
	}

	return newUser, token, nil
}

func (u *Usecase) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	uObj, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || uObj == nil {
		return nil, "", errs.NewUnauthorized("invalid email or password")
	}

	if !u.hasher.Verify(password, uObj.PasswordHash) {
		return nil, "", errs.NewUnauthorized("invalid email or password")
	}

	token, err := u.issuer.IssueToken(uObj)
	if err != nil {
		return nil, "", err
	}

	return uObj, token, nil
}

func (u *Usecase) GetMe(ctx context.Context, userID string) (*user.User, error) {
	uObj, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || uObj == nil {
		return nil, errs.NewNotFound("user", userID)
	}
	return uObj, nil
}

func (u *Usecase) UpdateMe(ctx context.Context, userID string, in UpdateMeInput) (*user.User, error) {
	uObj, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || uObj == nil {
		return nil, errs.NewNotFound("user", userID)
	}

	if in.Name != nil {
		uObj.Name = strings.TrimSpace(*in.Name)
	}
	if in.Phone != nil {
		uObj.Phone = strings.TrimSpace(*in.Phone)
	}
	if in.Address != nil {
		uObj.Address = strings.TrimSpace(*in.Address)
	}
	uObj.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, uObj); err != nil {
		return nil, err
	}

	return uObj, nil
}

func (u *Usecase) ChangePassword(ctx context.Context, userID, oldPass, newPass string) error {
	uObj, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || uObj == nil {
		return errs.NewNotFound("user", userID)
	}

	if !u.hasher.Verify(oldPass, uObj.PasswordHash) {
		return errs.NewValidation("incorrect old password", errs.FieldViolation{Field: "old_password", Issue: "does not match"})
	}

	if utf8.RuneCountInString(newPass) < 10 {
		return errs.NewValidation("new password too short", errs.FieldViolation{Field: "new_password", Issue: "must be at least 10 characters"})
	}

	newHash, err := u.hasher.Hash(newPass)
	if err != nil {
		return err
	}

	uObj.PasswordHash = newHash
	uObj.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, uObj)
}
