package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SessionStore manages authentication sessions.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

type Session struct {
	Token     string    `json:"token"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

var store = &SessionStore{
	sessions: make(map[string]*Session),
}

func init() {
	// Cleanup expired sessions every hour
	go func() {
		for range time.Tick(time.Hour) {
			store.CleanupExpired()
		}
	}()
}

// GenerateToken creates a random hex token for session identification.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession creates a new unverified session.
func (s *SessionStore) CreateSession(name, phone string) (*Session, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	sess := &Session{
		Token:     token,
		Name:      name,
		Phone:     phone,
		Verified:  false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Minute), // Sessions expire after 30 min if not verified
	}

	s.mu.Lock()
	s.sessions[token] = sess
	s.mu.Unlock()

	return sess, nil
}

// GetSession retrieves a session by token.
func (s *SessionStore) GetSession(token string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[token]
}

// MarkVerified marks a session as verified.
func (s *SessionStore) MarkVerified(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[token]
	if !ok {
		return false
	}
	sess.Verified = true
	sess.ExpiresAt = time.Now().Add(24 * time.Hour) // Verified sessions last 24 hours
	return true
}

// CleanupExpired removes expired sessions.
func (s *SessionStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for token, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, token)
		}
	}
}

// SendCodeRequest is the request body for starting phone verification.
type SendCodeRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// SendCodeResponse is returned after a session is created.
type SendCodeResponse struct {
	SessionToken string `json:"sessionToken"`
	Message      string `json:"message"`
}

// VerifyCodeRequest is the request body for verifying a code.
type VerifyCodeRequest struct {
	SessionToken string `json:"sessionToken"`
	Code         string `json:"code"`
}

// VerifyCodeResponse is the response after verifying a code.
type VerifyCodeResponse struct {
	Verified bool   `json:"verified"`
	Token    string `json:"token"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Message  string `json:"message"`
}

// SendCode handles POST /api/auth/send-code — creates a session (no outbound SMS).
func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Name == "" || req.Phone == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name and phone are required"})
		return
	}

	// Normalize phone number (strip non-digits, ensure E.164)
	phone := normalizePhone(req.Phone)

	sess, err := store.CreateSession(req.Name, phone)
	if err != nil {
		log.Printf("Session creation error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
		return
	}

	log.Printf("✅ Session created for %s (session: %s…)", phone, sess.Token[:8])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SendCodeResponse{
		SessionToken: sess.Token,
		Message:      "Session started. SMS is not used — submit any 4+ digit code to verify.",
	})
}

// VerifyCode handles POST /api/auth/verify-code — marks the session verified (no SMS provider).
func (h *Handler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.SessionToken == "" || req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session token and code are required"})
		return
	}

	// Get the session
	sess := store.GetSession(req.SessionToken)
	if sess == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session not found or expired"})
		return
	}

	if sess.Verified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(VerifyCodeResponse{
			Verified: true,
			Token:    sess.Token,
			Name:     sess.Name,
			Phone:    sess.Phone,
			Message:  "Already verified",
		})
		return
	}

	code := strings.TrimSpace(req.Code)
	if len(code) < 4 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Enter a verification code (at least 4 characters)."})
		return
	}

	// Mark session as verified
	store.MarkVerified(req.SessionToken)

	// Find or create patient in PostgreSQL
	ctx := r.Context()
	if _, err := h.pg.FindOrCreatePatient(ctx, sess.Name, sess.Phone, nil, nil); err != nil {
		log.Printf("FindOrCreatePatient error: %v", err)
	}

	log.Printf("✅ User verified: %s (%s)", sess.Name, sess.Phone)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(VerifyCodeResponse{
		Verified: true,
		Token:    sess.Token,
		Name:     sess.Name,
		Phone:    sess.Phone,
		Message:  "Phone verified successfully",
	})
}

// GetSession handles GET /api/auth/session
// It returns the current session info if valid.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "No session token"})
		return
	}

	// Strip "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	sess := store.GetSession(token)
	if sess == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session not found or expired"})
		return
	}

	if !sess.Verified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session not verified"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"verified": sess.Verified,
		"name":     sess.Name,
		"phone":    sess.Phone,
		"expires":  sess.ExpiresAt.Format(time.RFC3339),
	})
}

// normalizePhone strips non-digit characters and ensures E.164 format.
func normalizePhone(phone string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	if len(digits) == 10 {
		return "+1" + digits
	}
	if len(digits) == 11 && digits[0] == '1' {
		return "+" + digits
	}
	if strings.HasPrefix(phone, "+") {
		return phone
	}
	return "+" + digits
}
