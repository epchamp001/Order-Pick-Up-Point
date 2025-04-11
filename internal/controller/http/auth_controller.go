package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	http2 "order-pick-up-point/internal/service/http"
	"order-pick-up-point/pkg/validator"
)

type AuthController interface {
	DummyLogin(c *gin.Context)
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type authController struct {
	authSvc http2.AuthService
}

func NewAuthController(authSvc http2.AuthService) AuthController {
	return &authController{
		authSvc: authSvc,
	}
}

// DummyLogin godoc
// @Summary Dummy login for testing
// @Description Get a JWT token by passing a desired user role (client, employee, moderator) through dummy login.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.DummyLoginPostRequest true "Dummy login request with role"
// @Success 200 {object} dto.TokenResponse "JWT token"
// @Failure 400 {object} dto.Error "Invalid request body"
// @Failure 401 {object} dto.Error "Unauthorized: invalid role or error during dummy login"
// @Router /dummyLogin [post]
func (a *authController) DummyLogin(c *gin.Context) {
	var req dto.DummyLoginPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{
			Message: "invalid request body",
		})
		return
	}

	token, err := a.authSvc.DummyLogin(c, req.Role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		Token: token,
	})
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, password and role (client, employee, moderator).
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterPostRequest true "User registration data"
// @Success 201 {object} dto.RegisterResponse "User registration success response with user ID"
// @Failure 400 {object} dto.Error "Invalid request body"
// @Failure 500 {object} dto.Error "Internal server error"
// @Router /register [post]
func (a *authController) Register(c *gin.Context) {
	var req dto.RegisterPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{
			Message: "invalid request body",
		})
		return
	}

	if err := validator.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{
			Message: err.Error(),
		})
		return
	}

	if err := validator.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{
			Message: err.Error(),
		})
		return
	}

	userID, err := a.authSvc.Register(c, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.RegisterResponse{
		UserID: userID,
	})
}

// Login godoc
// @Summary Login a user
// @Description Login a user using email and password. Returns a JWT token if credentials are valid.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginPostRequest true "User login data"
// @Success 200 {object} dto.TokenResponse "JWT token"
// @Failure 400 {object} dto.Error "Invalid request body"
// @Failure 401 {object} dto.Error "Unauthorized: invalid credentials"
// @Router /login [post]
func (a *authController) Login(c *gin.Context) {
	var req dto.LoginPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{
			Message: "invalid request body",
		})
		return
	}

	token, err := a.authSvc.Login(c, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		Token: token,
	})
}
