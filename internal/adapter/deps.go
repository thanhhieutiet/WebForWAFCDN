// Package adapter implements interface ports defined by the usecase layer.
// It integrates external libraries for SQLite database access, JWT authentication,
// Argon2 password hashing, and WebSockets.
package adapter

import (
	_ "github.com/coder/websocket"
	_ "github.com/golang-jwt/jwt/v5"
	_ "golang.org/x/crypto/argon2"
	_ "modernc.org/sqlite"
)
