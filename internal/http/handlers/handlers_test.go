package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"cache-server/internal/cache/service"
	"cache-server/pkg/types"

	"github.com/gofiber/fiber/v3"
)

func setupApp() (*fiber.App, *Handlers) {
	app := fiber.New()
	
	cfg := &types.Config{
		MaxEntries: 1000,
		MaxMemory:  1024 * 1024,
		DefaultTTL: 60,
	}
	svc := service.NewService(cfg)
	h := NewHandlers(svc)

	app.Post("/api/v1/set", h.SetHandler)
	app.Get("/api/v1/get/:key", h.GetHandler)
	app.Delete("/api/v1/del/:key", h.DelHandler)
	app.Get("/api/v1/metrics", h.MetricsHandler)
	app.Get("/api/v1/health", h.HealthHandler)

	return app, h
}

func TestHealthHandler(t *testing.T) {
	app, _ := setupApp()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test health handler: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestSetAndGetHandler(t *testing.T) {
	app, _ := setupApp()

	// 1. Set a value
	setBody := SetRequest{
		Key:   "test_key",
		Value: json.RawMessage(`{"hello": "world"}`),
	}
	bodyBytes, _ := json.Marshal(setBody)
	reqSet := httptest.NewRequest("POST", "/api/v1/set", bytes.NewReader(bodyBytes))
	reqSet.Header.Set("Content-Type", "application/json")
	
	respSet, err := app.Test(reqSet)
	if err != nil {
		t.Fatalf("Failed to test set handler: %v", err)
	}
	if respSet.StatusCode != 200 {
		t.Errorf("Expected status code 200 for SET, got %d", respSet.StatusCode)
	}

	// 2. Get the value
	reqGet := httptest.NewRequest("GET", "/api/v1/get/test_key", nil)
	respGet, err := app.Test(reqGet)
	if err != nil {
		t.Fatalf("Failed to test get handler: %v", err)
	}
	if respGet.StatusCode != 200 {
		t.Errorf("Expected status code 200 for GET, got %d", respGet.StatusCode)
	}
}

func TestDelHandler(t *testing.T) {
	app, _ := setupApp()

	// Set value
	setBody := SetRequest{
		Key:   "del_key",
		Value: json.RawMessage(`"data"`),
	}
	bodyBytes, _ := json.Marshal(setBody)
	reqSet := httptest.NewRequest("POST", "/api/v1/set", bytes.NewReader(bodyBytes))
	reqSet.Header.Set("Content-Type", "application/json")
	app.Test(reqSet)

	// Delete value
	reqDel := httptest.NewRequest("DELETE", "/api/v1/del/del_key", nil)
	respDel, err := app.Test(reqDel)
	if err != nil {
		t.Fatalf("Failed to test del handler: %v", err)
	}
	if respDel.StatusCode != 200 {
		t.Errorf("Expected status code 200 for DEL, got %d", respDel.StatusCode)
	}

	// Get value should fail
	reqGet := httptest.NewRequest("GET", "/api/v1/get/del_key", nil)
	respGet, _ := app.Test(reqGet)
	if respGet.StatusCode != 404 {
		t.Errorf("Expected status code 404 for GET after DEL, got %d", respGet.StatusCode)
	}
}
