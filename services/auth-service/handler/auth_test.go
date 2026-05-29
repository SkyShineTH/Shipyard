package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SkyShineTH/shipyard/auth-service/handler"
	"github.com/SkyShineTH/shipyard/auth-service/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret-for-auth-handler-tests")
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func newRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	h := handler.NewAuthHandler(db)
	r.POST("/api/v1/register", h.Register)
	r.POST("/api/v1/login", h.Login)
	return r
}

func post(t *testing.T, router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRegister_Success(t *testing.T) {
	t.Parallel()
	rec := post(t, newRouter(newTestDB(t)), "/api/v1/register",
		map[string]string{"email": "new@example.com", "password": "password123"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body)
	}
}

func TestRegister_DuplicateEmail_Returns409(t *testing.T) {
	t.Parallel()
	r := newRouter(newTestDB(t))
	body := map[string]string{"email": "dup@example.com", "password": "password123"}
	post(t, r, "/api/v1/register", body)
	rec := post(t, r, "/api/v1/register", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body)
	}
}

func TestRegister_InvalidEmail_Returns400(t *testing.T) {
	t.Parallel()
	rec := post(t, newRouter(newTestDB(t)), "/api/v1/register",
		map[string]string{"email": "not-an-email", "password": "password123"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body)
	}
}

func TestRegister_ShortPassword_Returns400(t *testing.T) {
	t.Parallel()
	rec := post(t, newRouter(newTestDB(t)), "/api/v1/register",
		map[string]string{"email": "a@example.com", "password": "short"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body)
	}
}

func TestLogin_ValidCredentials_ReturnsToken(t *testing.T) {
	t.Parallel()
	r := newRouter(newTestDB(t))
	post(t, r, "/api/v1/register", map[string]string{"email": "user@example.com", "password": "password123"})
	rec := post(t, r, "/api/v1/login", map[string]string{"email": "user@example.com", "password": "password123"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Fatal("expected non-empty token in response body")
	}
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
	t.Parallel()
	r := newRouter(newTestDB(t))
	post(t, r, "/api/v1/register", map[string]string{"email": "user@example.com", "password": "password123"})
	rec := post(t, r, "/api/v1/login", map[string]string{"email": "user@example.com", "password": "wrongpassword"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body)
	}
}

func TestLogin_UnknownEmail_Returns401(t *testing.T) {
	t.Parallel()
	rec := post(t, newRouter(newTestDB(t)), "/api/v1/login",
		map[string]string{"email": "nobody@example.com", "password": "password123"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body)
	}
}

// Login must return the same error message whether the email is unknown or the
// password is wrong — leaking which one is true would allow account enumeration.
func TestLogin_DoesNotLeakEmailExistence(t *testing.T) {
	t.Parallel()
	r := newRouter(newTestDB(t))
	post(t, r, "/api/v1/register", map[string]string{"email": "exists@example.com", "password": "password123"})

	recWrongPw := post(t, r, "/api/v1/login", map[string]string{"email": "exists@example.com", "password": "wrongpassword"})
	recUnknown := post(t, r, "/api/v1/login", map[string]string{"email": "nobody@example.com", "password": "wrongpassword"})

	if recWrongPw.Body.String() != recUnknown.Body.String() {
		t.Fatalf("response body differs between wrong-password and unknown-email: %q vs %q",
			recWrongPw.Body.String(), recUnknown.Body.String())
	}
}
