package transaction_middleware

import (
	"github.com/Montheankul-K/jod-jod/domains/transaction"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

type ITransactionMiddleware interface {
	SetTxnPagination(next echo.HandlerFunc) echo.HandlerFunc
	SetPeriodFilter(next echo.HandlerFunc) echo.HandlerFunc
	SetGetAllTxnFilter(next echo.HandlerFunc) echo.HandlerFunc
	SetGetByTxnTypeRequest(next echo.HandlerFunc) echo.HandlerFunc
	SetGetByCategoryRequest(next echo.HandlerFunc) echo.HandlerFunc
}

type transactionMiddleware struct {
	logger echo.Logger
}

func NewTransactionMiddleware(logger echo.Logger) ITransactionMiddleware {
	return &transactionMiddleware{logger: logger}
}

func (m *transactionMiddleware) SetTxnPagination(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var err error
		req := transaction.Pagination{
			PageItem: 5,
			Page:     1,
		}

		pageItem := c.QueryParam("page-item")
		if pageItem != "" {
			req.PageItem, err = strconv.Atoi(pageItem)
			if err != nil {
				m.logger.Error(err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "internal server error",
				})
			}
		}

		page := c.QueryParam("page")
		if page != "" {
			req.Page, err = strconv.Atoi(page)
			if err != nil {
				m.logger.Error(err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "internal server error",
				})
			}
		}
		c.Set("pagination", req)
		return next(c)
	}
}

func (m *transactionMiddleware) SetPeriodFilter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var startDate, endDate time.Time
		var err error
		startDateStr := c.QueryParam("start-date")
		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				m.logger.Error(err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "internal server error",
				})
			} else {
				startDate = startDate.In(time.Local).Add(time.Hour * 0).Add(time.Minute * 0).Add(time.Second * 0)
			}
		}

		endDateStr := c.QueryParam("end-date")
		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				m.logger.Error(err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "internal server error",
				})
			} else {
				endDate = endDate.In(time.Local).Add(time.Hour * 23).Add(time.Minute * 59).Add(time.Second * 59)
			}
		}
		req := transaction.PeriodFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		c.Set("filter", req)
		return next(c)
	}
}

func (m *transactionMiddleware) SetGetAllTxnFilter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var date time.Time
		var err error
		dateStr := c.QueryParam("date")
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				m.logger.Error(err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "internal server error",
				})
			}
		}

		category := c.QueryParam("category")
		txnType := c.QueryParam("txn-type")
		req := transaction.GetAllTxnFilter{
			Date:     &date,
			Category: category,
			TxnType:  txnType,
		}
		c.Set("filter", req)
		return next(c)
	}
}

func (m *transactionMiddleware) SetGetByTxnTypeRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		spenderIdStr := c.Param("spender-id")
		if spenderIdStr == "" {
			m.logger.Error("spender id is empty")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "spender id is required",
			})
		}

		txnType := c.QueryParam("txn-type")
		if txnType == "" {
			m.logger.Error("transaction type is empty")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "transaction type is required",
			})
		}

		spenderId, err := strconv.ParseUint(spenderIdStr, 10, 64)
		if err != nil {
			m.logger.Error(err)
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "spender id is invalid",
			})
		}

		req := transaction.GetByTxnTypeRequest{
			SpenderId: uint(spenderId),
			TxnType:   txnType,
		}
		c.Set("", req)
		return next(c)
	}
}

func (m *transactionMiddleware) SetGetByCategoryRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		spenderIdStr := c.Param("spender-id")
		if spenderIdStr == "" {
			m.logger.Error("spender id is empty")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "spender id is required",
			})
		}

		category := c.QueryParam("category")
		if category == "" {
			m.logger.Error("category is empty")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "category is required",
			})
		}

		txnType := c.QueryParam("txn-type")
		if txnType == "" {
			m.logger.Error("transaction type is empty")
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "transaction type is required",
			})
		}

		spenderId, err := strconv.ParseUint(spenderIdStr, 10, 64)
		if err != nil {
			m.logger.Error(err)
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "spender id is invalid",
			})
		}

		req := transaction.GetByCategoryRequest{
			SpenderId: uint(spenderId),
			Category:  category,
			TxnType:   txnType,
		}
		c.Set("", req)
		return next(c)
	}
}
