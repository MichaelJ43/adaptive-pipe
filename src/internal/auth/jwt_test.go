package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSignParseRoundTrip(t *testing.T) {
	secret := []byte("test-secret-at-least-32-bytes-long!!")
	tid := uuid.New()
	uid := uuid.New()
	tok, err := SignJWT(secret, tid, uid, "admin", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseJWT(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.TenantID != tid || c.UserID != uid || c.Username != "admin" || c.Role != "admin" {
		t.Fatalf("claims mismatch %+v", c)
	}
}
