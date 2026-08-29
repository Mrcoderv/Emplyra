package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type EmployeeHandler struct {
	svc *services.EmployeeService
}

func NewEmployeeHandler(svc *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) List(c *gin.Context) {
	var params dto.EmployeeListParams
	_ = c.ShouldBindQuery(&params)
	p := utils.NewPagination(params.Page, params.PageSize)
	items, total, err := h.svc.List(middleware.TenantID(c), p, struct{ Search, DepartmentID, DesignationID, ManagerID, Status, EmploymentType, SortBy, SortOrder string }{
		Search: params.Search, DepartmentID: params.DepartmentID, DesignationID: params.DesignationID,
		ManagerID: params.ManagerID, Status: params.Status, EmploymentType: params.EmploymentType,
		SortBy: params.SortBy, SortOrder: params.SortOrder,
	})
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "employees", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *EmployeeHandler) Get(c *gin.Context) {
	e, err := h.svc.Get(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "employee", e)
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.EmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	e, err := h.svc.Create(middleware.TenantID(c), employeeInputFromRequest(req), actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "employee created", e)
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	var req dto.EmployeeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	e, err := h.svc.Update(middleware.TenantID(c), c.Param("id"), employeeInputFromUpdate(req), actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "employee updated", e)
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	if err := h.svc.Delete(middleware.TenantID(c), c.Param("id"), actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "employee deleted", nil)
}

func (h *EmployeeHandler) MyProfile(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	e, err := h.svc.GetByUserID(p.UserID)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "my employee profile", e)
}

func employeeInputFromRequest(req dto.EmployeeRequest) services.EmployeeInput {
	return services.EmployeeInput{
		EmployeeCode: req.EmployeeCode, FirstName: req.FirstName, LastName: req.LastName,
		Email: req.Email, Phone: req.Phone, DateOfBirth: req.DateOfBirth, Gender: req.Gender,
		Address: req.Address, EmergencyContact: req.EmergencyContact, JoiningDate: req.JoiningDate,
		EmploymentType: req.EmploymentType, Status: req.Status, DepartmentID: req.DepartmentID,
		DesignationID: req.DesignationID, ManagerID: req.ManagerID, UserID: req.UserID,
	}
}

func employeeInputFromUpdate(req dto.EmployeeUpdateRequest) services.EmployeeInput {
	return services.EmployeeInput{
		EmployeeCode: req.EmployeeCode, FirstName: req.FirstName, LastName: req.LastName,
		Email: req.Email, Phone: req.Phone, DateOfBirth: req.DateOfBirth, Gender: req.Gender,
		Address: req.Address, EmergencyContact: req.EmergencyContact, JoiningDate: req.JoiningDate,
		EmploymentType: req.EmploymentType, Status: req.Status, DepartmentID: req.DepartmentID,
		DesignationID: req.DesignationID, ManagerID: req.ManagerID, UserID: req.UserID,
	}
}
