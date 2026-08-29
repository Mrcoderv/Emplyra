package handlers

import (
	"os"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type DocumentHandler struct {
	svc *services.DocumentService
	emp *repositories.EmployeeRepository
}

func NewDocumentHandler(svc *services.DocumentService, emp *repositories.EmployeeRepository) *DocumentHandler {
	return &DocumentHandler{svc: svc, emp: emp}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	employeeID := c.PostForm("employee_id")
	if employeeID == "" {
		responses.Error(c, 400, "employee_id is required", nil)
		return
	}
	if p.Role == string(models.RoleEmployee) {
		own, err := h.emp.FindByUserID(p.UserID)
		if err != nil || own.ID != employeeID {
			responses.Error(c, 403, "forbidden", nil)
			return
		}
	}
	title := c.PostForm("title")
	if title == "" {
		responses.Error(c, 400, "title is required", nil)
		return
	}
	uploadedBy := utils.CloneString(p.UserID)
	fh, err := c.FormFile("file")
	if err != nil {
		responses.Error(c, 400, "file is required", nil)
		return
	}
	doc, err := h.svc.Upload(employeeID, title, c.PostForm("type"), uploadedBy, fh)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "document uploaded", doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	employeeID := c.Query("employee_id")
	principal := middleware.GetPrincipal(c)
	if principal != nil && principal.Role == string(models.RoleEmployee) {
		if emp, err := h.emp.FindByUserID(principal.UserID); err == nil {
			employeeID = emp.ID
		}
	}
	items, total, err := h.svc.List(p, employeeID, c.Query("type"), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "documents", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *DocumentHandler) Get(c *gin.Context) {
	d, err := h.svc.Get(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "document", d)
}

func (h *DocumentHandler) Download(c *gin.Context) {
	path, _, err := h.svc.DownloadPath(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	if _, err := os.Stat(path); err != nil {
		responses.Error(c, 404, "file not found", nil)
		return
	}
	c.File(path)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	if err := h.svc.Delete(c.Param("id"), utils.CloneString(p.UserID)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "document deleted", nil)
}

// NotificationHandler reads and updates the caller's own notifications.
type NotificationHandler struct {
	svc interface {
		List(userID string, unreadOnly bool, p utils.Pagination) ([]models.Notification, int64, error)
		UnreadCount(userID string) (int64, error)
		MarkRead(userID, id string) error
		MarkAllRead(userID string) error
	}
}

func NewNotificationHandler(svc interface {
	List(userID string, unreadOnly bool, p utils.Pagination) ([]models.Notification, int64, error)
	UnreadCount(userID string) (int64, error)
	MarkRead(userID, id string) error
	MarkAllRead(userID string) error
}) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	pg := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.List(p.UserID, c.Query("unread") == "true", pg)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "notifications", responses.List{Items: items, Total: total, Page: pg.Page, PageSize: pg.PageSize, TotalPages: pg.TotalPages(total)})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	n, err := h.svc.UnreadCount(p.UserID)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "unread notifications", gin.H{"unread": n})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	if err := h.svc.MarkRead(p.UserID, c.Param("id")); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "notification marked read", nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	if err := h.svc.MarkAllRead(p.UserID); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "all notifications marked read", nil)
}
