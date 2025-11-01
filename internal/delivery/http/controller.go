package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	uuid "github.com/satori/go.uuid"
	"github.com/th1enq/es-demo/internal/command"
	"github.com/th1enq/es-demo/internal/dto"
	"github.com/th1enq/es-demo/internal/mappers"
	"github.com/th1enq/es-demo/internal/query"
	"github.com/th1enq/es-demo/internal/service"
	"github.com/th1enq/es-demo/pkg/constants"
)

type Controller struct {
	BankAccountService *service.BankAccountService
	validator          *validator.Validate
}

func NewController(
	bankAccountService *service.BankAccountService,
) *Controller {
	return &Controller{
		BankAccountService: bankAccountService,
		validator:          validator.New(),
	}
}

// CreateBankAccount godoc
// @Summary      Create Bank Account
// @Description  Create a new bank account
// @Tags         BankAccount
// @Accept       json
// @Produce      json
// @Param        request  body      command.CreateBankAccountCommand  true  "Create Bank Account Request"
// @Success      200      {object}  dto.APIResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /api/v1/bank_accounts [post]
func (b *Controller) CreateBankAccount(c *gin.Context) {
	var command command.CreateBankAccountCommand

	if err := c.ShouldBindBodyWithJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	command.AggregateID = uuid.NewV4().String()

	if err := b.validator.StructCtx(c, command); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	if err := b.BankAccountService.Commands.CreateBankAccount.Handle(
		c,
		command,
	); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to create bank account",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"bank account created successfully",
		nil,
	))
}

// DepositeBalance godoc
// @Summary      Deposite Balance
// @Description  Deposite balance to bank account
// @Tags         BankAccount
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Bank Account ID"
// @Param        request  body      command.DepositeBalanceCommand    true  "Deposite Balance Request"
// @Success      200      {object}  dto.APIResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /api/v1/bank_accounts/{id}/deposite [post]
func (b *Controller) DepositeBalance(c *gin.Context) {
	var command command.DepositeBalanceCommand

	if err := c.ShouldBindBodyWithJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	command.AggregateID = c.Param(constants.ID)

	if err := b.validator.StructCtx(c, command); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	if err := b.BankAccountService.Commands.DepositeBalance.Handle(
		c,
		command,
	); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to deposite balance",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"balance deposited successfully",
		nil,
	))
}

// WithdrawBalance godoc
// @Summary      Withdraw Balance
// @Description  Withdraw balance from bank account
// @Tags         BankAccount
// @Accept       json
// @Produce      json
// @Param        id       path      string                         true  "Bank Account ID"
// @Param        request  body      command.WithdrawBalanceCommand     true  "Withdraw Balance Request"
// @Success      200      {object}  dto.APIResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /api/v1/bank_accounts/{id}/withdraw [post]
func (b *Controller) WithdrawBalance(c *gin.Context) {
	var command command.WithdrawBalanceCommand

	if err := c.ShouldBindBodyWithJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	command.AggregateID = c.Param(constants.ID)

	if err := b.validator.StructCtx(c, command); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request body",
			err.Error(),
		))
		return
	}

	if err := b.BankAccountService.Commands.WithdrawBalance.Handle(
		c,
		command,
	); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to withdraw balance",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"balance withdrawn successfully",
		nil,
	))
}

// GetBankAccountByID godoc
// @Summary      Get Bank Account by ID
// @Description  Retrieve bank account details by ID
// @Tags         BankAccount
// @Accept       json
// @Produce      json
// @Param        id               path      string  true  "Bank Account ID"
// @Param        from_event_store query     string  false "Fetch data from event store (true/false)"
// @Success      200              {object}  dto.APIResponse
// @Failure      400              {object}  dto.APIResponse
// @Failure      500              {object}  dto.APIResponse
// @Router       /api/v1/bank_accounts/{id} [get]
func (b *Controller) GetBankAccountByID(c *gin.Context) {
	var query query.GetBankAccountByIDQuery

	query.AggregateID = c.Param(constants.ID)

	fromEventStore := c.Query("from_event_store")
	if fromEventStore != "" {
		isFromStore, err := strconv.ParseBool(fromEventStore)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
				dto.CodeBadRequest,
				"invalid from_event_store query parameter",
				err.Error(),
			))
			return
		}
		query.FromEventStore = isFromStore
	}

	if err := b.validator.StructCtx(c, query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request",
			err.Error(),
		))
		return
	}

	result, err := b.BankAccountService.Query.GetBankAccountByID.Handle(
		c,
		query,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to get bank account",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"bank account retrieved successfully",
		mappers.BankAccountMongoProjectionToHttp(result),
	))
}

// GetEventsHistory godoc
// @Summary      Get Events History
// @Description  Retrieve all events for a specific bank account
// @Tags         BankAccount
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Bank Account ID"
// @Success      200  {object}  dto.APIResponse
// @Failure      400  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/bank_accounts/{id}/events [get]
func (b *Controller) GetEventsHistory(c *gin.Context) {
	var query query.GetEventsHistoryQuery

	query.AggregateID = c.Param(constants.ID)

	if err := b.validator.StructCtx(c, query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request",
			err.Error(),
		))
		return
	}

	result, err := b.BankAccountService.Query.GetEventsHistory.Handle(c, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to get events history",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"events history retrieved successfully",
		result,
	))
}
