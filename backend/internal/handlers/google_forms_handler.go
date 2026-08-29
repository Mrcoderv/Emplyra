package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type GoogleFormHandler struct {
	svc *services.GoogleFormService
}

func NewGoogleFormHandler(svc *services.GoogleFormService) *GoogleFormHandler {
	return &GoogleFormHandler{svc: svc}
}

func (h *GoogleFormHandler) Connect(c *gin.Context) {
	req, ok := bindGoogleForm(c)
	if !ok {
		return
	}
	a := middleware.MustPrincipal(c)
	integ, err := h.svc.Connect(middleware.TenantID(c), c.Param("id"), services.GoogleFormInput{
		FormURL: req.GoogleFormURL, SheetID: req.GoogleSheetID, SheetName: req.ResponseSheet,
		HeaderRow: req.HeaderRow, Status: req.Status, StatusDetail: req.StatusDetail,
		FieldMapping: toFieldMap(req.FieldMapping),
	}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "google form connected", integ)
}

func (h *GoogleFormHandler) Get(c *gin.Context) {
	integ, err := h.svc.Get(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google form integration", integ)
}

func (h *GoogleFormHandler) Update(c *gin.Context) {
	req, ok := bindGoogleForm(c)
	if !ok {
		return
	}
	a := middleware.MustPrincipal(c)
	integ, err := h.svc.Update(middleware.TenantID(c), c.Param("id"), services.GoogleFormInput{
		FormURL: req.GoogleFormURL, SheetID: req.GoogleSheetID, SheetName: req.ResponseSheet,
		HeaderRow: req.HeaderRow, Status: req.Status, StatusDetail: req.StatusDetail,
		FieldMapping: toFieldMap(req.FieldMapping),
	}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google form integration updated", integ)
}

func (h *GoogleFormHandler) Disconnect(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.Disconnect(middleware.TenantID(c), c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google form disconnected", nil)
}

func (h *GoogleFormHandler) Sync(c *gin.Context) {
	var req dto.GoogleFormSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	if req.Mode == "" {
		req.Mode = "incremental"
	}
	if req.Mode != "incremental" && req.Mode != "full" {
		responses.Error(c, 400, "validation failed", map[string]string{"mode": "must be 'incremental' or 'full'"})
		return
	}
	a := middleware.MustPrincipal(c)
	result, err := h.svc.Sync(middleware.TenantID(c), c.Param("id"), req.Mode, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google form synchronized", result)
}

func (h *GoogleFormHandler) SyncStatus(c *gin.Context) {
	integ, counts, err := h.svc.SyncStatus(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google form sync status", gin.H{
		"integration": integ,
		"counts":      counts,
	})
}

func (h *GoogleFormHandler) Responses(c *gin.Context) {
	var q dto.GoogleFormsResponseQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	p := utils.NewPagination(q.Page, q.PageSize)
	items, total, err := h.svc.Responses(middleware.TenantID(c), c.Param("id"), q.Status, p)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google form responses", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *GoogleFormHandler) OAuthAuthorize(c *gin.Context) {
	authURL, state, err := h.svc.OAuthAuthorize()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "google oauth authorize", dto.GoogleOAuthAuthorizeResponse{AuthURL: authURL, State: state})
}

// OAuthCallback is reached by the browser after the user consents on Google.
func (h *GoogleFormHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		responses.Error(c, http.StatusBadRequest, "oauth callback missing code or state", nil)
		return
	}
	redirect, err := h.svc.OAuthCallback(code, state)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	c.Redirect(http.StatusFound, redirect)
}

func bindGoogleForm(c *gin.Context) (dto.GoogleFormRequest, bool) {
	var req dto.GoogleFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return req, false
	}
	return req, true
}

func toFieldMap(dtos []dto.GoogleFieldMap) []services.FieldMap {
	if len(dtos) == 0 {
		return nil
	}
	out := make([]services.FieldMap, 0, len(dtos))
	for _, d := range dtos {
		out = append(out, services.FieldMap{Source: d.Source, Target: d.Target})
	}
	return out
}
