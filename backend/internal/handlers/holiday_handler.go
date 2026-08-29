package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type HolidayHandler struct {
	svc *services.HolidayService
}

func NewHolidayHandler(svc *services.HolidayService) *HolidayHandler {
	return &HolidayHandler{svc: svc}
}

func (h *HolidayHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Query("from"), c.Query("to"), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "holidays", items)
}

func (h *HolidayHandler) Get(c *gin.Context) {
	item, err := h.svc.Get(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "holiday", item)
}

func (h *HolidayHandler) Create(c *gin.Context) {
	var req dto.HolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	item, err := h.svc.Create(struct{ Name, Date, Description, Type, Status string }{
		Name: sanitizeField(req.Name), Date: req.Date, Description: req.Description, Type: req.Type, Status: req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "holiday created", item)
}

func (h *HolidayHandler) Update(c *gin.Context) {
	var req dto.HolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	item, err := h.svc.Update(c.Param("id"), struct{ Name, Date, Description, Type, Status string }{
		Name: sanitizeField(req.Name), Date: req.Date, Description: req.Description, Type: req.Type, Status: req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "holiday updated", item)
}

func (h *HolidayHandler) Delete(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	if err := h.svc.Delete(c.Param("id"), actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "holiday deleted", nil)
}
