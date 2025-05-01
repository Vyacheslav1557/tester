package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"time"
)

type Session struct {
	Id        string    `json:"id"`
	UserId    int32     `json:"user_id"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UserAgent string    `json:"user_agent"`
	Ip        string    `json:"ip"`
}

func (s *Session) Valid() error {
	if uuid.Validate(s.Id) != nil {
		return errors.New("invalid session id")
	}
	if s.UserId == 0 {
		return errors.New("empty user id")
	}
	if s.CreatedAt.IsZero() {
		return errors.New("empty created at")
	}
	if s.ExpiresAt.IsZero() {
		return errors.New("empty expires at")
	}
	if s.UserAgent == "" {
		return errors.New("empty user agent")
	}
	if s.Ip == "" {
		return errors.New("empty ip")
	}
	return nil
}

func (s *Session) JSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Session) UserIdHash() string {
	return sha256string(strconv.FormatInt(int64(s.UserId), 10))
}

func (s *Session) SessionIdHash() string {
	return sha256string(s.Id)
}

func (s *Session) Key() string {
	return fmt.Sprintf("userid:%s:sessionid:%s", s.UserIdHash(), s.SessionIdHash())
}
func sha256string(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

type JWT struct {
	SessionId string `json:"session_id"`
	UserId    int32  `json:"user_id"`
	Role      Role   `json:"role"`
	IssuedAt  int64  `json:"iat"`
}

func (j JWT) Valid() error {
	if uuid.Validate(j.SessionId) != nil {
		return errors.New("invalid session id")
	}
	if j.UserId == 0 {
		return errors.New("empty user id")
	}
	if j.IssuedAt == 0 {
		return errors.New("empty issued at")
	}
	return nil
}

type Credentials struct {
	Username string
	Password string
}

type Device struct {
	Ip       string
	UseAgent string
}
