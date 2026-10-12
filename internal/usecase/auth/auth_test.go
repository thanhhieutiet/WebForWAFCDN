package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

type fakeUserRepo struct {
	users map[string]*user.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*user.User)}
}

func (f *fakeUserRepo) Create(ctx context.Context, u *user.User) error {
	if _, ok := f.users[u.ID]; ok {
		return errors.New("user ID already exists")
	}
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*user.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	for _, u := range f.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeUserRepo) Update(ctx context.Context, u *user.User) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserRepo) List(ctx context.Context, limit, offset int) ([]user.User, int, error) {
	var list []user.User
	for _, u := range f.users {
		list = append(list, *u)
	}
	return list, len(list), nil
}

type fakeHasher struct{}

func (h fakeHasher) Hash(p string) (string, error) {
	return "hashed:" + p, nil
}

func (h fakeHasher) Verify(p, hash string) bool {
	return "hashed:"+p == hash
}

type fakeIssuer struct{}

func (i fakeIssuer) IssueToken(u *user.User) (string, error) {
	return "token:" + u.ID, nil
}

func (i fakeIssuer) ValidateToken(t string) (*port.Claims, error) {
	parts := strings.Split(t, ":")
	if len(parts) == 2 && parts[0] == "token" {
		return &port.Claims{UserID: parts[1], Role: user.RoleCustomer, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	return nil, errors.New("invalid token")
}

type fakeIDGen struct {
	next string
}

func (g fakeIDGen) NewID() string {
	if g.next != "" {
		return g.next
	}
	return "u-123"
}

func TestAuthUsecase(t *testing.T) {
	repo := newFakeUserRepo()
	hasher := fakeHasher{}
	issuer := fakeIssuer{}
	idGen := fakeIDGen{next: "user-1"}

	uc := New(repo, hasher, issuer, idGen)
	ctx := context.Background()

	// Register short password
	_, _, err := uc.Register(ctx, RegisterInput{Email: "test@example.com", Password: "short"})
	if err == nil {
		t.Fatalf("expected error for short password")
	}

	// Successful register
	u, token, err := uc.Register(ctx, RegisterInput{
		Email:    "test@example.com",
		Password: "SecurePassword123!",
		Name:     "Test User",
	})
	if err != nil {
		t.Fatalf("unexpected error registering: %v", err)
	}
	if u.ID != "user-1" || token != "token:user-1" {
		t.Fatalf("invalid user or token: %+v, %s", u, token)
	}

	// Register duplicate email
	_, _, err = uc.Register(ctx, RegisterInput{
		Email:    "test@example.com",
		Password: "SecurePassword123!",
	})
	if err == nil {
		t.Fatalf("expected conflict error on duplicate email")
	}

	// Login wrong password
	_, _, err = uc.Login(ctx, "test@example.com", "WrongPassword!")
	if err == nil {
		t.Fatalf("expected error on wrong password")
	}

	// Login success
	uLog, tokenLog, err := uc.Login(ctx, "test@example.com", "SecurePassword123!")
	if err != nil || uLog.ID != "user-1" || tokenLog == "" {
		t.Fatalf("login failed: %v", err)
	}

	// GetMe
	me, err := uc.GetMe(ctx, "user-1")
	if err != nil || me.Email != "test@example.com" {
		t.Fatalf("GetMe failed: %v", err)
	}

	// UpdateMe
	newName := "Updated Name"
	updated, err := uc.UpdateMe(ctx, "user-1", UpdateMeInput{Name: &newName})
	if err != nil || updated.Name != "Updated Name" {
		t.Fatalf("UpdateMe failed: %v", err)
	}

	// ChangePassword wrong old password
	err = uc.ChangePassword(ctx, "user-1", "wrong", "NewSecurePassword123!")
	if err == nil {
		t.Fatalf("expected error changing password with bad old password")
	}

	// ChangePassword success
	err = uc.ChangePassword(ctx, "user-1", "SecurePassword123!", "NewSecurePassword123!")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Login with new password
	_, _, err = uc.Login(ctx, "test@example.com", "NewSecurePassword123!")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}

	// Non-existent GetMe
	_, err = uc.GetMe(ctx, "non-existent")
	if err == nil {
		t.Fatalf("expected error for non-existent user in GetMe")
	}

	// Non-existent UpdateMe
	_, err = uc.UpdateMe(ctx, "non-existent", UpdateMeInput{})
	if err == nil {
		t.Fatalf("expected error for non-existent user in UpdateMe")
	}

	// Non-existent ChangePassword
	err = uc.ChangePassword(ctx, "non-existent", "old", "new")
	if err == nil {
		t.Fatalf("expected error for non-existent user in ChangePassword")
	}

	// Short new password in ChangePassword
	err = uc.ChangePassword(ctx, "user-1", "NewSecurePassword123!", "short")
	if err == nil {
		t.Fatalf("expected error for short new password in ChangePassword")
	}
}
