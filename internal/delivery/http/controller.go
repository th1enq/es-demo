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
	ReplayService      *service.ReplayService
	validator          *validator.Validate
}

func NewController(
	bankAccountService *service.BankAccountService,
	replayService *service.ReplayService,
) *Controller {
	return &Controller{
		BankAccountService: bankAccountService,
		ReplayService:      replayService,
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

// GetBankAccountByVersion godoc
// @Summary      Get Bank Account by Version
// @Description  Retrieve bank account state at a specific version
// @Tags         BankAccount
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Bank Account ID"
// @Param        version path      string  true  "Version number"
// @Success      200     {object}  dto.APIResponse
// @Failure      400     {object}  dto.APIResponse
// @Failure      500     {object}  dto.APIResponse
// @Router       /api/v1/bank_accounts/{id}/version/{version} [get]
func (b *Controller) GetBankAccountByVersion(c *gin.Context) {
	var query query.GetBankAccountByVersionQuery

	query.AggregateID = c.Param(constants.ID)

	versionStr := c.Param("version")
	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid version parameter",
			err.Error(),
		))
		return
	}
	query.Version = version

	if err := b.validator.StructCtx(c, query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"invalid request",
			err.Error(),
		))
		return
	}

	result, err := b.BankAccountService.Query.GetBankAccountByVersion.Handle(
		c,
		query,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to get bank account by version",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"bank account at version retrieved successfully",
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

// ReplayAllEvents godoc
// @Summary      Replay All Events to Elasticsearch
// @Description  Replay all events from PostgreSQL event store to Elasticsearch for demonstration of Event Sourcing replay capability
// @Tags         Replay
// @Accept       json
// @Produce      json
// @Param        recreate_index query  string  false "Whether to recreate the Elasticsearch index (true/false)"
// @Success      200            {object}  dto.APIResponse
// @Failure      400            {object}  dto.APIResponse
// @Failure      500            {object}  dto.APIResponse
// @Router       /api/v1/replay/events [post]
func (b *Controller) ReplayAllEvents(c *gin.Context) {
	recreateIndexStr := c.Query("recreate_index")
	recreateIndex := false
	if recreateIndexStr != "" {
		parsed, err := strconv.ParseBool(recreateIndexStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
				dto.CodeBadRequest,
				"invalid recreate_index query parameter",
				err.Error(),
			))
			return
		}
		recreateIndex = parsed
	}

	result, err := b.ReplayService.ReplayAllEvents(c, recreateIndex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to replay events",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"events replayed successfully to Elasticsearch",
		result,
	))
}

// GetAccountFromElasticsearch godoc
// @Summary      Get Bank Account from Elasticsearch
// @Description  Retrieve bank account details from Elasticsearch (replayed data)
// @Tags         Replay
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Bank Account ID"
// @Success      200  {object}  dto.APIResponse
// @Failure      400  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/replay/accounts/{id} [get]
func (b *Controller) GetAccountFromElasticsearch(c *gin.Context) {
	aggregateID := c.Param(constants.ID)

	if aggregateID == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.CodeBadRequest,
			"aggregate ID is required",
			"missing aggregate ID in path",
		))
		return
	}

	result, err := b.ReplayService.GetAccountByID(c, aggregateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to get account from Elasticsearch",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"account retrieved successfully from Elasticsearch",
		result,
	))
}

// SearchAccountsInElasticsearch godoc
// @Summary      Search Bank Accounts in Elasticsearch
// @Description  Search for bank accounts in Elasticsearch with various filters
// @Tags         Replay
// @Accept       json
// @Produce      json
// @Param        email     query     string  false "Email filter"
// @Param        firstName query     string  false "First name filter"
// @Param        lastName  query     string  false "Last name filter"
// @Param        status    query     string  false "Account status filter"
// @Success      200       {object}  dto.APIResponse
// @Failure      400       {object}  dto.APIResponse
// @Failure      500       {object}  dto.APIResponse
// @Router       /api/v1/replay/accounts/search [get]
func (b *Controller) SearchAccountsInElasticsearch(c *gin.Context) {
	// Build Elasticsearch query from query parameters
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
			},
		},
		"size": 100, // Limit results
	}

	mustClauses := []interface{}{}

	// Add filters based on query parameters
	if email := c.Query("email"); email != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"email": email,
			},
		})
	}

	if firstName := c.Query("firstName"); firstName != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"firstName": firstName,
			},
		})
	}

	if lastName := c.Query("lastName"); lastName != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"lastName": lastName,
			},
		})
	}

	if status := c.Query("status"); status != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"status": status,
			},
		})
	}

	// If no filters provided, use match_all
	if len(mustClauses) == 0 {
		query["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	} else {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustClauses
	}

	results, err := b.ReplayService.SearchAccounts(c, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to search accounts in Elasticsearch",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"accounts search completed successfully",
		map[string]interface{}{
			"accounts": results,
			"total":    len(results),
		},
	))
}

// GetAccountSummary godoc
// @Summary      Get Account Analytics Summary
// @Description  Get analytics summary of all bank accounts from Elasticsearch
// @Tags         Replay
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/replay/summary [get]
func (b *Controller) GetAccountSummary(c *gin.Context) {
	summary, err := b.ReplayService.GetAccountSummary(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to get account summary",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"account summary retrieved successfully",
		summary,
	))
}

// DeleteElasticsearchIndex godoc
// @Summary      Delete Elasticsearch Index
// @Description  Delete the bank accounts index from Elasticsearch (for demo purposes)
// @Tags         Replay
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /api/v1/replay/index [delete]
func (b *Controller) DeleteElasticsearchIndex(c *gin.Context) {
	// Delete the bank accounts index
	err := b.ReplayService.DeleteElasticsearchIndex(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.CodeInternalServerError,
			"failed to delete Elasticsearch index",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.CodeSuccess,
		"Elasticsearch index deleted successfully",
		nil,
	))
}
