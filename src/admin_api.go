package goule

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type APIHandler func(*RouteRequest, []byte) ([]byte, error)

type AuthAPIBody struct {
	Password string `json:"password"`
}

func APIHandlers() map[string]APIHandler {
	return map[string]APIHandler{"auth": AuthAPI}
}

func AuthAPI(req *RouteRequest, body []byte) ([]byte, error) {
	var value AuthAPIBody
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, err
	}

	// Check the password
	hash := sha256.Sum256([]byte(value.Password))
	hex := hex.EncodeToString(hash[:])
	adminHash := req.Overseer.GetAdminSettings().PasswordHash
	if strings.ToLower(hex) != strings.ToLower(adminHash) {
		return nil, errors.New("The provided password was incorrect.")
	}

	// Create a new session
	sessionId := req.Overseer.GetSessions().Login()
	cookie := &http.Cookie{Name: SessionIdCookie, Value: sessionId,
		Path: req.AdminRule.Path}
	http.SetCookie(req.Response, cookie)
	return []byte("\"Authentication successful.\""), nil
}
