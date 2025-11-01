package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	uuid "github.com/satori/go.uuid"
	"github.com/th1enq/es-demo/internal/dto"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"github.com/th1enq/es-demo/internal/service"
	"go.uber.org/zap"
)

type AuthController struct {
	authService service.AuthService
	validator   *validator.Validate
	logger      *zap.Logger
}

func NewAuthController(
	authService service.AuthService,
	logger *zap.Logger,
) *AuthController {
	return &AuthController{
		authService: authService,
		validator:   validator.New(),
		logger:      logger,
	}
}

// Login godoc
// @Summary      User Login
// @Description  Authenticate user and return JWT tokens
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "Login Request"
// @Success      200      {object}  dto.APIResponse{data=dto.LoginResponse}
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /api/v1/auth/login [post]
func (a *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid login request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	if err := a.validator.StructCtx(c, req); err != nil {
		a.logger.Error("Login request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeValidationError,
			"validation failed",
			err.Error(),
		))
		return
	}

	response, err := a.authService.Login(c, req)
	if err != nil {
		a.logger.Error("Login failed", zap.String("email", req.Email), zap.Error(err))

		if err == bankAccountErrors.ErrInvalidCredentials ||
			err == bankAccountErrors.ErrAccountInactive {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				dto.CodeUnauthorized,
				"authentication failed",
				err.Error(),
			))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"login failed",
			err.Error(),
		))
		return
	}

	a.logger.Info("Login successful", zap.String("user_id", response.User.ID))
	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"login successful",
		response,
	))
}

// Register godoc
// @Summary      User Registration
// @Description  Register a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "Registration Request"
// @Success      201      {object}  dto.APIResponse{data=dto.RegisterResponse}
// @Failure      400      {object}  dto.APIResponse
// @Failure      409      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /api/v1/auth/register [post]
func (a *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid registration request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	// Generate ID if not provided
	if req.ID == "" {
		req.ID = uuid.NewV4().String()
	}

	if err := a.validator.StructCtx(c, req); err != nil {
		a.logger.Error("Registration request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeValidationError,
			"validation failed",
			err.Error(),
		))
		return
	}

	response, err := a.authService.Register(c, req)
	if err != nil {
		a.logger.Error("Registration failed", zap.String("email", req.Email), zap.Error(err))

		if err == bankAccountErrors.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(
				dto.CodeConflict,
				"registration failed",
				err.Error(),
			))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"registration failed",
			err.Error(),
		))
		return
	}

	a.logger.Info("Registration successful", zap.String("user_id", response.UserID))
	c.JSON(http.StatusCreated, dto.NewSuccessResponse(
		dto.CodeCreated,
		"registration successful",
		response,
	))
}

// RefreshToken godoc
// @Summary      Refresh JWT Token
// @Description  Refresh access token using refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshTokenRequest  true  "Refresh Token Request"
// @Success      200      {object}  dto.APIResponse{data=dto.LoginResponse}
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /api/v1/auth/refresh [post]
func (a *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid refresh token request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	if err := a.validator.StructCtx(c, req); err != nil {
		a.logger.Error("Refresh token request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeValidationError,
			"validation failed",
			err.Error(),
		))
		return
	}

	response, err := a.authService.RefreshToken(c, req.RefreshToken)
	if err != nil {
		a.logger.Error("Token refresh failed", zap.Error(err))

		if err == bankAccountErrors.ErrInvalidToken ||
			err == bankAccountErrors.ErrAccountInactive {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				dto.CodeTokenExpired,
				"token refresh failed",
				err.Error(),
			))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"token refresh failed",
			err.Error(),
		))
		return
	}

	a.logger.Info("Token refresh successful", zap.String("user_id", response.User.ID))
	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"token refreshed successfully",
		response,
	))
}

// Logout godoc
// @Summary      User Logout
// @Description  Logout user (client-side token removal)
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.APIResponse
// @Router       /api/v1/auth/logout [post]
// @Security     BearerAuth
func (a *AuthController) Logout(c *gin.Context) {
	// In a stateless JWT implementation, logout is typically handled client-side
	// by removing the token from storage. However, we can log the event.
	userID, exists := c.Get("user_id")
	if exists {
		a.logger.Info("User logged out", zap.Any("user_id", userID))
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"logout successful",
		nil,
	))
}
