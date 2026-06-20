package transport

import (
	"fmt"

	"go-modular-cqrs-monolith/modules/order/internal/command"
	"go-modular-cqrs-monolith/modules/order/internal/query"
	"go-modular-cqrs-monolith/platform/httpx/response"

	"github.com/gofiber/fiber/v3"
)

type OrderHTTPHandler struct {
	cmd *command.Command
	qry *query.Query
}

func NewOrderHTTPHandler(cmd *command.Command, qry *query.Query) *OrderHTTPHandler {
	return &OrderHTTPHandler{
		cmd: cmd,
		qry: qry,
	}
}

func (h *OrderHTTPHandler) GetByID(c fiber.Ctx) error {
	var params struct {
		OrderID string `uri:"order_id"`
	}
	if err := c.Bind().URI(&params); err != nil {
		return err
	}

	res, err := h.qry.GetById(c, params.OrderID)
	if err != nil {
		return fmt.Errorf("qry.GetById: %w", err)
	}

	return response.OK(c, res)
}
