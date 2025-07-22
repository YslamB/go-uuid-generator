package http

import (
	"fmt"
	"id-generator/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	uc *usecase.IDUsecase
}

func NewHandler(uc *usecase.IDUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {

	app.Get("/generate", func(c *fiber.Ctx) error {
		id, err := h.uc.GenerateID()

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid count parameter: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{"id": id})
	})

	app.Get("/generate/:count", func(c *fiber.Ctx) error {
		count := c.Params("count")

		var countInt int

		if _, err := fmt.Sscanf(count, "%d", &countInt); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid count parameter",
			})
		}

		if countInt <= 0 || countInt > 10000 {
			return c.Status(400).JSON(fiber.Map{
				"error": "count must be between 1 and 10000",
			})
		}

		ids, duration, err := h.uc.GenerateIDs(countInt)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid count parameter: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"count":       len(*ids),
			"duration_ms": duration,
			"ids":         ids,
		})
	})
}
