package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/mapper"
	http2 "order-pick-up-point/internal/service/http"
	"strconv"
	"strings"
	"time"
)

type PvzController interface {
	CreatePvz(c *gin.Context)
	GetPvzsInfo(c *gin.Context)
	CreateReception(c *gin.Context)
	AddProduct(c *gin.Context)
	DeleteLastProduct(c *gin.Context)
	CloseReception(c *gin.Context)
	GetPvzsInfoOptimized(c *gin.Context)
}

type pvzController struct {
	pvzSvc http2.PvzService
}

func NewPvzController(pvzSvc http2.PvzService) PvzController {
	return &pvzController{
		pvzSvc: pvzSvc,
	}
}

// CreatePvz godoc
// @Summary Create a new PVZ
// @Security BearerAuth
// @Description Create a new PVZ. Only users with a moderator role can create a PVZ.
// @Tags pvz
// @Accept json
// @Produce json
// @Param request body dto.CreatePvzPostRequest true "PVZ creation data"
// @Success 201 {object} dto.CreatePvzResponse "Newly created PVZ information"
// @Failure 400 {object} dto.Error "Invalid request body"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /pvz [post]
func (p *pvzController) CreatePvz(c *gin.Context) {
	if !CheckRole(c, "moderator") {
		return
	}

	var req dto.CreatePvzPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid request body"})
		return
	}

	pvzId, err := p.pvzSvc.CreatePvz(c, req.City)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreatePvzResponse{PvzId: pvzId})
}

// GetPvzsInfo godoc
// @Summary Get PVZ information
// @Security BearerAuth
// @Description Retrieve a paginated list of PVZ with reception and product details.
// @Tags pvz
// @Accept json
// @Produce json
// @Param page query int true "Page number" example(1)
// @Param limit query int true "Number of items per page" example(10)
// @Param startDate query string false "Filter: start date in RFC3339 format" example("2025-04-09T00:00:00Z")
// @Param endDate query string false "Filter: end date in RFC3339 format" example("2025-04-09T23:59:59Z")
// @Success 200 {array} dto.PvzGet200ResponseInner "List of PVZ information"
// @Failure 400 {object} dto.Error "Invalid query parameters"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /pvz [get]
func (p *pvzController) GetPvzsInfo(c *gin.Context) {
	if !CheckRole(c, "moderator", "employee") {
		return
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	if pageStr == "" || limitStr == "" {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "missing page or limit parameter"})
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid page parameter"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid limit parameter"})
		return
	}

	var startDate, endDate *time.Time
	startDateStr := c.Query("startDate")
	if startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid startDate format"})
			return
		}
		startDate = &t
	}
	endDateStr := c.Query("endDate")
	if endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid endDate format"})
			return
		}
		endDate = &t
	}

	pvzInfos, err := p.pvzSvc.GetPvzsInfo(c, page, limit, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: fmt.Sprintf("failed to get PVZ info: %v", err)})
		return
	}

	response := make([]dto.PvzGet200ResponseInner, 0, len(pvzInfos))
	for _, info := range pvzInfos {
		response = append(response, mapper.PvzInfoEntityToResponse(info))
	}

	c.JSON(http.StatusOK, response)
}

// CreateReception godoc
// @Summary Create a new reception for goods
// @Security BearerAuth
// @Description Initiate a new reception for a specified PVZ. Only employees can create a reception.
// @Tags pvz
// @Accept json
// @Produce json
// @Param request body dto.CreateReceptionRequest true "Reception creation data"
// @Success 201 {object} dto.CreateReceptionResponse "Reception created, returning its ID"
// @Failure 400 {object} dto.Error "Invalid request body"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /receptions [post]
func (p *pvzController) CreateReception(c *gin.Context) {
	if !CheckRole(c, "employee") {
		return
	}

	var req dto.CreateReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid request body"})
		return
	}

	receptionTime := req.DateTime
	if receptionTime.IsZero() {
		receptionTime = time.Now()
	}

	receptionID, err := p.pvzSvc.CreateReception(c, req.PvzId, receptionTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateReceptionResponse{ReceptionId: receptionID})
}

// AddProduct godoc
// @Summary Add a product to the current reception
// @Security BearerAuth
// @Description Add a new product to the last open reception for a given PVZ. Only employees can add products.
// @Tags pvz
// @Accept json
// @Produce json
// @Param request body dto.ProductsPostRequest true "Product addition data"
// @Success 201 {object} dto.ProductsPostResponse "Product added, returning its ID"
// @Failure 400 {object} dto.Error "Invalid request body"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /products [post]
func (p *pvzController) AddProduct(c *gin.Context) {
	if !CheckRole(c, "employee") {
		return
	}

	var req dto.ProductsPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid request body"})
		return
	}

	productID, err := p.pvzSvc.AddProduct(c, req.PvzId, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ProductsPostResponse{ProductId: productID})
}

// DeleteLastProduct godoc
// @Summary Delete the last added product from the current reception
// @Security BearerAuth
// @Description Delete the last product that was added to an open reception for a given PVZ (LIFO order). Only employees can delete products.
// @Tags pvz
// @Accept json
// @Produce json
// @Param pvzId path string true "PVZ ID" example("pvz123")
// @Success 200 {object} dto.DeleteProductResponse "Product deletion success message"
// @Failure 400 {object} dto.Error "Bad request: missing pvzId"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /pvz/{pvzId}/delete_last_product [post]
func (p *pvzController) DeleteLastProduct(c *gin.Context) {
	if !CheckRole(c, "employee") {
		return
	}

	pvzId := c.Param("pvzId")
	if pvzId == "" {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "pvzId parameter is required"})
		return
	}

	if err := p.pvzSvc.DeleteLastProduct(c, pvzId); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.DeleteProductResponse{Message: "product deleted successfully"})
}

// CloseReception godoc
// @Summary Close the current reception
// @Security BearerAuth
// @Description Close the last open reception for a specified PVZ, finalizing the reception process. Only employees can close receptions.
// @Tags pvz
// @Accept json
// @Produce json
// @Param pvzId path string true "PVZ ID" example("pvz123")
// @Success 200 {object} dto.CloseReceptionResponse "Closed reception details with its ID"
// @Failure 400 {object} dto.Error "Bad request: missing pvzId"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /pvz/{pvzId}/close_last_reception [post]
func (p *pvzController) CloseReception(c *gin.Context) {
	if !CheckRole(c, "employee") {
		return
	}

	pvzId := c.Param("pvzId")
	if pvzId == "" {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "pvzId parameter is required"})
		return
	}

	receptionID, err := p.pvzSvc.CloseReception(c, pvzId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.CloseReceptionResponse{ReceptionId: receptionID})
}

func CheckRole(c *gin.Context, allowedRoles ...string) bool {
	roleVal, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error{Message: "missing role in token"})
		c.Abort()
		return false
	}

	roleStr, ok := roleVal.(string)
	if !ok {
		c.JSON(http.StatusForbidden, dto.Error{Message: "invalid role type"})
		c.Abort()
		return false
	}

	roleStr = strings.ToLower(roleStr)
	for _, allowed := range allowedRoles {
		if roleStr == strings.ToLower(allowed) {
			return true
		}
	}

	c.JSON(http.StatusForbidden, dto.Error{Message: fmt.Sprintf("access denied, allowed roles: %v", allowedRoles)})
	c.Abort()
	return false
}

// GetPvzsInfoOptimized godoc
// @Summary Get PVZ info with receptions and products (optimized)
// @Security BearerAuth
// @Description Optimized method to retrieve a paginated list of PVZ with all related receptions and products.
// @Tags pvz
// @Accept json
// @Produce json
// @Param page query int true "Page number" example(1)
// @Param limit query int true "Number of items per page" example(10)
// @Param startDate query string false "Filter: start date in RFC3339 format" example("2025-04-09T00:00:00Z")
// @Param endDate query string false "Filter: end date in RFC3339 format" example("2025-04-09T23:59:59Z")
// @Success 200 {array} dto.PvzGet200ResponseInner "List of PVZ information (optimized)"
// @Failure 400 {object} dto.Error "Invalid query parameters"
// @Failure 401 {object} dto.Error "Unauthorized: missing token or insufficient privileges"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /pvz/optimized [get]
func (p *pvzController) GetPvzsInfoOptimized(c *gin.Context) {
	if !CheckRole(c, "moderator", "employee") {
		return
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	if pageStr == "" || limitStr == "" {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "missing page or limit parameter"})
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid page parameter"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid limit parameter"})
		return
	}

	var startDate, endDate *time.Time
	startDateStr := c.Query("startDate")
	if startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid startDate format"})
			return
		}
		startDate = &t
	}
	endDateStr := c.Query("endDate")
	if endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.Error{Message: "invalid endDate format"})
			return
		}
		endDate = &t
	}

	pvzInfos, err := p.pvzSvc.GetPvzsInfoOptimized(c, page, limit, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: fmt.Sprintf("failed to get optimized PVZ info: %v", err)})
		return
	}

	response := make([]dto.PvzGet200ResponseInner, 0, len(pvzInfos))
	for _, info := range pvzInfos {
		response = append(response, mapper.PvzInfoEntityToResponse(info))
	}

	c.JSON(http.StatusOK, response)
}
