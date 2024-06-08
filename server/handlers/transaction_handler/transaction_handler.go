package transaction_handler

import (
	"errors"
	"fmt"
	"github.com/Montheankul-K/jod-jod/domains/transaction"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type ITransactionHandler interface {
	SaveByManual(c echo.Context) error
	GetDetails(c echo.Context) error
	GetSummary(c echo.Context) error
	GetBalance(c echo.Context) error
	GetByCategory(c echo.Context) error
	GetByPeriod(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
	GetAllTxn(c echo.Context) error
}

type transactionHandler struct {
	transactionService transaction.ITransactionService
	logger             echo.Logger
}

func NewTransactionHandler(transactionService transaction.ITransactionService, logger echo.Logger) ITransactionHandler {
	return &transactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

func (h *transactionHandler) SaveByManual(c echo.Context) error {
	var req transaction.Transaction
	if err := c.Bind(&req); err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"message": "request body is invalid",
		})
	}

	validate := validator.New()
	err := validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	result, err := h.transactionService.SaveByManual(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusCreated, echo.Map{"spender_id": result})
}

func (h *transactionHandler) GetDetails(c echo.Context) error {
	req := c.Get("").(transaction.GetByTxnTypeRequest)
	validate := validator.New()
	err := validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	result, err := h.transactionService.GetDetails(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "transaction not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *transactionHandler) GetSummary(c echo.Context) error {
	req := c.Get("").(transaction.GetByTxnTypeRequest)
	validate := validator.New()
	err := validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	result, err := h.transactionService.GetSummary(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "transaction not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *transactionHandler) GetBalance(c echo.Context) error {
	spenderIdStr := c.Param("spender-id")
	if spenderIdStr == "" {
		h.logger.Error("spender-id is empty")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "spender id is required"})
	}

	spenderId, err := strconv.ParseUint(spenderIdStr, 10, 64)
	if err != nil {
		h.logger.Error("spender-id is invalid")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "spender id is invalid"})
	}

	result, err := h.transactionService.GetBalance(uint(spenderId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "transaction not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *transactionHandler) GetByCategory(c echo.Context) error {
	req := c.Get("").(transaction.GetByCategoryRequest)
	validate := validator.New()
	err := validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	result, err := h.transactionService.GetByCategory(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "transaction not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *transactionHandler) GetByPeriod(c echo.Context) error {
	req := c.Get("").(transaction.GetByTxnTypeRequest)
	validate := validator.New()
	err := validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	filter := c.Get("filter").(transaction.PeriodFilter)
	result, err := h.transactionService.GetByPeriod(req, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "transaction not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *transactionHandler) Update(c echo.Context) error {
	var req transaction.Transaction
	if err := c.Bind(&req); err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"message": "request body is invalid",
		})
	}

	validate := validator.New()
	err := validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	err = h.transactionService.Update(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": fmt.Sprintf("update transaction with transaction id: %d success", req.ID)})
}

func (h *transactionHandler) Delete(c echo.Context) error {
	spenderIdStr := c.Param("spender-id")
	if spenderIdStr == "" {
		h.logger.Error("spender-id is empty")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "spender id is required"})
	}

	txnIdStr := c.Param("txn-id")
	if txnIdStr == "" {
		h.logger.Error("txn-id is empty")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "txn-id is required"})
	}

	spenderId, err := strconv.ParseUint(spenderIdStr, 10, 64)
	if err != nil {
		h.logger.Error("spender-id is invalid")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "spender id is invalid"})
	}

	txnId, err := strconv.ParseUint(txnIdStr, 10, 64)
	if err != nil {
		h.logger.Error("txn-id is invalid")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "txn-id is invalid"})
	}

	err = h.transactionService.Delete(uint(spenderId), uint(txnId))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": fmt.Sprintf("delete transaction with transaction id: %d success", txnId)})
}

func (h *transactionHandler) GetAllTxn(c echo.Context) error {
	filter := c.Get("filter").(transaction.GetAllTxnFilter)
	pagination := c.Get("pagination").(transaction.Pagination)
	result, err := h.transactionService.GetAllTxn(filter, pagination)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "transaction not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err})
	}
	return c.JSON(http.StatusOK, result)
}
