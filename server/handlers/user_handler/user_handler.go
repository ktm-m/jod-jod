package user_handler

import (
	"errors"
	"fmt"
	"github.com/Montheankul-K/jod-jod/domains/user"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type IUserHandler interface {
	GetUsers(c echo.Context) error
	GetUser(c echo.Context) error
	CreateUser(c echo.Context) error
	Login(c echo.Context) error
	RegenToken(c echo.Context) error
	UpdateInfo(c echo.Context) error
	UpdatePassword(c echo.Context) error
	DeleteUser(c echo.Context) error
}

type userHandler struct {
	userService user.IUserService
	logger      echo.Logger
}

func NewUserHandler(userService user.IUserService, logger echo.Logger) IUserHandler {
	return &userHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *userHandler) GetUsers(c echo.Context) error {
	pagination := c.Get("pagination").(user.Pagination)
	result, err := h.userService.GetUsers(pagination)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"message": "user not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *userHandler) GetUser(c echo.Context) error {
	userIdStr := c.Param("user-id")
	if userIdStr == "" {
		h.logger.Error("result id is empty")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "result id is required",
		})
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		h.logger.Error("user id is invalid")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "user id is invalid",
		})
	}

	result, err := h.userService.GetUser(uint(userId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"message": "user not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *userHandler) CreateUser(c echo.Context) error {
	var req user.Users
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
			"message": errors.New("request body is invalid").Error(),
		})
	}

	result, err := h.userService.CreateUser(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"user_id": result,
	})
}

func (h *userHandler) Login(c echo.Context) error {
	var req user.LoginRequest
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
			"message": errors.New("request body is invalid").Error(),
		})
	}

	result, err := h.userService.Login(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "user is invalid",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *userHandler) RegenToken(c echo.Context) error {
	var req user.RegenTokenRequest
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

	result, err := h.userService.RegenToken(req)
	if err != nil {
		if errors.Is(err, errors.New("failed to parse user id")) || errors.Is(err, errors.New("failed to parse user id")) {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": err,
			})
		}
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"access_token": result,
	})
}

func (h *userHandler) UpdateInfo(c echo.Context) error {
	userIdStr := c.Param("user-id")
	if userIdStr == "" {
		h.logger.Error("user id is empty")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "user id is required",
		})
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		h.logger.Error("user id is invalid")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "user id is invalid",
		})
	}

	var req user.UpdateInfoRequest
	if err = c.Bind(&req); err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"message": "request body is invalid",
		})
	}

	validate := validator.New()
	err = validate.Struct(&req)
	if err != nil {
		h.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err,
		})
	}

	newReq := user.Users{
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Email:     req.Email,
	}
	err = h.userService.UpdateInfo(uint(userId), newReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": fmt.Sprintf("update user info for user id: %d successfully", userId),
	})
}

func (h *userHandler) UpdatePassword(c echo.Context) error {
	var req user.UpdatePasswordRequest
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
			"message": errors.New("request body is invalid").Error(),
		})
	}

	err = h.userService.UpdatePassword(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": fmt.Sprintf("update password for user id: %d successfully", req.ID),
	})
}

func (h *userHandler) DeleteUser(c echo.Context) error {
	userIdStr := c.Param("user-id")
	if userIdStr == "" {
		h.logger.Error("user id is empty")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "user id is required",
		})
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		h.logger.Error("user id is invalid")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "user id is invalid",
		})
	}

	err = h.userService.DeleteUser(uint(userId))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err,
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": fmt.Sprintf("delete user for user id: %d successfully", userId),
	})
}
