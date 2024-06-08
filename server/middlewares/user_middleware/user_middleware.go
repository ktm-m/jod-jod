package user_middleware

import (
	"errors"
	"github.com/Montheankul-K/jod-jod/config"
	"github.com/Montheankul-K/jod-jod/domains/user"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type IUserMiddleware interface {
	SetUserPagination(next echo.HandlerFunc) echo.HandlerFunc
	ValidateToken(next echo.HandlerFunc) echo.HandlerFunc
}

type userMiddleware struct {
	cfg    *config.Config
	logger echo.Logger
}

func NewUserMiddleware(cfg *config.Config, logger echo.Logger) IUserMiddleware {
	return &userMiddleware{
		cfg:    cfg,
		logger: logger,
	}
}

func (m *userMiddleware) SetUserPagination(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var err error
		req := user.Pagination{
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

func (m *userMiddleware) ValidateToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		secret := m.cfg.Auth.Secret
		auth := c.Request().Header.Get("Authorization")
		if auth == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "authorization header is missing",
			})
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "authorization header is invalid",
			})
		}

		tokenString := parts[1]
		claims := &user.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil {
			var validationErr *jwt.ValidationError
			if errors.As(err, &validationErr) {
				switch {
				case validationErr.Errors&jwt.ValidationErrorMalformed != 0:
					return c.JSON(http.StatusBadRequest, echo.Map{
						"message": "invalid token format",
					})
				case validationErr.Errors&jwt.ValidationErrorExpired != 0:
					return c.JSON(http.StatusUnauthorized, echo.Map{
						"message": "token is expired",
					})
				case validationErr.Errors&jwt.ValidationErrorNotValidYet != 0:
					return c.JSON(http.StatusUnauthorized, echo.Map{
						"message": "token is not valid",
					})
				case validationErr.Errors&jwt.ValidationErrorSignatureInvalid != 0:
					return c.JSON(http.StatusUnauthorized, echo.Map{
						"message": "signature is invalid",
					})
				default:
					return c.JSON(http.StatusUnauthorized, echo.Map{
						"message": "token is invalid",
					})
				}
			}
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "internal server error",
			})
		}

		if !token.Valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "invalid token",
			})
		}

		if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "token is expired",
			})
		}
		c.Set("token", tokenString)
		c.Set("claims", claims)
		return next(c)
	}
}
