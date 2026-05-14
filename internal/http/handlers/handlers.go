// Package handlers provides HTTP handlers for the cache server.
package handlers

import (
	"encoding/json"

	"cache-server/internal/cache/service"

	"github.com/gofiber/fiber/v3"
)

type Handlers struct {
	s *service.Service
}

func NewHandlers(s *service.Service) *Handlers {
	return &Handlers{s: s}
}

type SetRequest struct {
	Key   string          `json:"key"`
	TTL   uint64          `json:"ttl,omitempty"`
	Value json.RawMessage `json:"value"`
}

func (h *Handlers) SetHandler(c fiber.Ctx) error {
	var req SetRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if req.Key == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Key required"})
	}
	if len(req.Value) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Value required"})
	}

	if err := h.s.Set(req.Key, req.Value, req.TTL); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handlers) GetHandler(c fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Key required"})
	}

	value, ok := h.s.Get(key)
	if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "Key not found"})
	}

	// Return as JSON raw for zero-copy
	return c.JSON(fiber.Map{"value": json.RawMessage(value)})
}

func (h *Handlers) DelHandler(c fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Key required"})
	}

	h.s.Delete(key)
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handlers) MetricsHandler(c fiber.Ctx) error {
	return c.JSON(h.s.Metrics())
}

func (h *Handlers) HealthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "healthy"})
}
