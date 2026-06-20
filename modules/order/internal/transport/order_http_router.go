package transport

import "github.com/gofiber/fiber/v3"

type Middlewares struct {
	Auth    fiber.Handler
	Limiter fiber.Handler
}

func Mount(c *fiber.App, handler *OrderHTTPHandler, middlewares Middlewares) {
	g := c.Group("/api/v1/orders")
	g.Use(
		middlewares.Limiter,
		middlewares.Auth,
	)

	g.Get("/:order_id", handler.GetByID)
}
