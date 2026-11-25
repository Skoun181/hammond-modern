package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/akhilrex/hammond/db"
)

const AccessLifetime = 15 * time.Minute
const RefreshLifetime = 20 * 24 * time.Hour

var sessionDB *sql.DB

type SessionTokens struct {
	Access  string
	Refresh string
}

func OpenSessions(config string) error {
	if len(os.Getenv("JWT_SECRET")) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 bytes")
	}
	path, err := filepath.Abs(filepath.Join(config, "hammond-sessions.db"))
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	address := url.URL{Scheme: "file", Path: path, RawQuery: "_busy_timeout=5000&_txlock=immediate"}
	conn, err := sql.Open("sqlite3", address.String())
	if err != nil {
		return err
	}
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		access TEXT PRIMARY KEY, refresh TEXT NOT NULL UNIQUE, user_id TEXT NOT NULL,
		password_binding TEXT NOT NULL, access_until INTEGER NOT NULL, refresh_until INTEGER NOT NULL
	)`)
	if err != nil {
		conn.Close()
		return err
	}
	sessionDB = conn
	return nil
}

func sessionDigest(value string) string {
	mac := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func sessionSecret() (string, error) {
	var material [32]byte
	if _, err := rand.Read(material[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(material[:]), nil
}

func insertSession(tx *sql.Tx, user *db.User) (SessionTokens, error) {
	access, err := sessionSecret()
	if err != nil {
		return SessionTokens{}, err
	}
	refresh, err := sessionSecret()
	if err != nil {
		return SessionTokens{}, err
	}
	now := time.Now()
	if _, err = tx.Exec("DELETE FROM sessions WHERE refresh_until <= ?", now.Unix()); err != nil {
		return SessionTokens{}, err
	}
	_, err = tx.Exec("INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?)",
		sessionDigest(access), sessionDigest(refresh), user.ID, sessionDigest(user.Password),
		now.Add(AccessLifetime).Unix(), now.Add(RefreshLifetime).Unix())
	return SessionTokens{Access: access, Refresh: refresh}, err
}

func NewSession(user *db.User) (SessionTokens, error) {
	tx, err := sessionDB.Begin()
	if err != nil {
		return SessionTokens{}, err
	}
	defer tx.Rollback()
	tokens, err := insertSession(tx, user)
	if err != nil {
		return SessionTokens{}, err
	}
	return tokens, tx.Commit()
}

func sessionUser(id, binding string) (*db.User, error) {
	var user db.User
	if err := db.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if user.IsDisabled || !hmac.Equal([]byte(binding), []byte(sessionDigest(user.Password))) {
		return nil, fmt.Errorf("session is no longer valid")
	}
	return &user, nil
}

func AccessSession(token string) (*db.User, error) {
	if len(token) != 43 || sessionDB == nil {
		return nil, sql.ErrNoRows
	}
	var id, binding string
	err := sessionDB.QueryRow("SELECT user_id, password_binding FROM sessions WHERE access = ? AND access_until > ?",
		sessionDigest(token), time.Now().Unix()).Scan(&id, &binding)
	if err != nil {
		return nil, err
	}
	return sessionUser(id, binding)
}

func RotateSession(token string) (*db.User, SessionTokens, error) {
	if len(token) != 43 || sessionDB == nil {
		return nil, SessionTokens{}, sql.ErrNoRows
	}
	tx, err := sessionDB.Begin()
	if err != nil {
		return nil, SessionTokens{}, err
	}
	defer tx.Rollback()
	var id, binding string
	digest := sessionDigest(token)
	err = tx.QueryRow("SELECT user_id, password_binding FROM sessions WHERE refresh = ? AND refresh_until > ?",
		digest, time.Now().Unix()).Scan(&id, &binding)
	if err != nil {
		return nil, SessionTokens{}, err
	}
	user, err := sessionUser(id, binding)
	if err != nil {
		return nil, SessionTokens{}, err
	}
	if _, err := tx.Exec("DELETE FROM sessions WHERE refresh = ?", digest); err != nil {
		return nil, SessionTokens{}, err
	}
	tokens, err := insertSession(tx, user)
	if err != nil {
		return nil, SessionTokens{}, err
	}
	return user, tokens, tx.Commit()
}

func RevokeSession(access, refresh string) error {
	if sessionDB == nil {
		return nil
	}
	_, err := sessionDB.Exec("DELETE FROM sessions WHERE access = ? OR refresh = ?", sessionDigest(access), sessionDigest(refresh))
	return err
}

func RevokeUserSessions(userID string) error {
	_, err := sessionDB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}
