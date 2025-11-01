package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/dto"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"go.uber.org/zap"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
}

type authService struct {
	queryService QueryService
	commandBus   CommandBus
	jwtSecret    []byte
	logger       *zap.Logger
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(
	queryService QueryService,
	commandBus CommandBus,
	jwtSecret string,
	logger *zap.Logger,
) AuthService {
	return &authService{
		queryService: queryService,
		commandBus:   commandBus,
		jwtSecret:    []byte(jwtSecret),
		logger:       logger,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	s.logger.Info("Processing login request", zap.String("email", req.Email))

	// Get bank account by email
	bankAccount, err := s.queryService.GetBankAccountByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to get bank account by email", zap.Error(err))
		return nil, bankAccountErrors.ErrInvalidCredentials
	}

	// Check password
	if err := bankAccount.CheckPassword(req.Password); err != nil {
		s.logger.Warn("Invalid password attempt", zap.String("email", req.Email))
		return nil, bankAccountErrors.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(bankAccount)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(bankAccount)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Login successful", zap.String("user_id", bankAccount.AggregateID))

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: dto.UserInfo{
			ID:        bankAccount.AggregateID,
			Email:     bankAccount.Email,
			FirstName: bankAccount.FirstName,
			LastName:  bankAccount.LastName,
		},
	}, nil
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	s.logger.Info("Processing registration request", zap.String("email", req.Email))

	// Check if email already exists
	_, err := s.queryService.GetBankAccountByEmail(ctx, req.Email)
	if err == nil {
		s.logger.Warn("Registration attempt with existing email", zap.String("email", req.Email))
		return nil, bankAccountErrors.ErrEmailAlreadyExists
	}

	// Create bank account command
	createCmd := dto.CreateBankAccountRequest{
		AggregateID: req.ID,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Balance:     req.InitialBalance,
		Status:      "active",
		Password:    req.Password,
	}

	// Execute command
	if err := s.commandBus.CreateBankAccount(ctx, createCmd); err != nil {
		s.logger.Error("Failed to create bank account", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Registration successful", zap.String("user_id", req.ID))

	return &dto.RegisterResponse{
		UserID:  req.ID,
		Email:   req.Email,
		Message: "Account created successfully",
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	// Validate refresh token
	token, err := s.ValidateToken(refreshToken)
	if err != nil || !token.Valid {
		return nil, bankAccountErrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, bankAccountErrors.ErrInvalidToken
	}

	// Get current user data
	bankAccount, err := s.queryService.GetBankAccountByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	accessToken, err := s.generateAccessToken(bankAccount)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(bankAccount)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: dto.UserInfo{
			ID:        bankAccount.AggregateID,
			Email:     bankAccount.Email,
			FirstName: bankAccount.FirstName,
			LastName:  bankAccount.LastName,
		},
	}, nil
}

func (s *authService) generateAccessToken(bankAccount *domain.BankAccount) (string, error) {
	claims := CustomClaims{
		UserID: bankAccount.AggregateID,
		Email:  bankAccount.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   bankAccount.AggregateID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *authService) generateRefreshToken(bankAccount *domain.BankAccount) (string, error) {
	claims := CustomClaims{
		UserID: bankAccount.AggregateID,
		Email:  bankAccount.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   bankAccount.AggregateID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
