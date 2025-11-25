package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func setupApp() *fiber.App {
	app := fiber.New()
	InitializeRoute(app)
	return app
}

func TestHealthHandler(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if len(body) == 0 {
		t.Fatalf("expected non-empty body")
	}
}

func TestWebHookTelegram(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if len(body) == 0 {
		t.Fatalf("expected non-empty body")
	}
}
