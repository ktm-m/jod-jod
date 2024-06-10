package server

import (
	"github.com/Montheankul-K/jod-jod/domains/transaction"
	"github.com/Montheankul-K/jod-jod/domains/user"
	"github.com/Montheankul-K/jod-jod/repository/transaction_repository"
	"github.com/Montheankul-K/jod-jod/repository/user_repository"
	"github.com/Montheankul-K/jod-jod/server/handlers/health"
	"github.com/Montheankul-K/jod-jod/server/handlers/transaction_handler"
	"github.com/Montheankul-K/jod-jod/server/handlers/user_handler"
	"github.com/Montheankul-K/jod-jod/server/middlewares/transaction_middleware"
	"github.com/Montheankul-K/jod-jod/server/middlewares/user_middleware"
)

func (s *server) healthCheckRouter() {
	router := s.app.Group("/v1/healths")
	healthHandler := health.NewHealthHandler(s.cfg.Server)

	router.GET("/health-check", healthHandler.HealthCheck)
}

func (s *server) userRouter() {
	router := s.app.Group("/v1/users")
	userMiddleware := user_middleware.NewUserMiddleware(s.cfg, s.app.Logger)

	userRepository := user_repository.NewUserRepository(s.db.Connect(), s.app.Logger, s.redisClient)
	userService := user.NewUserService(userRepository, s.app.Logger)
	userHandler := user_handler.NewUserHandler(userService, s.app.Logger)

	router.GET("/get", userHandler.GetUsers, userMiddleware.ValidateToken, userMiddleware.SetUserPagination)
	router.GET("/get/:user-id", userHandler.GetUser, userMiddleware.ValidateToken)
	router.POST("/create", userHandler.CreateUser)
	router.POST("/login", userHandler.Login)
	router.POST("/regen-token", userHandler.RegenToken)
	router.PUT("/update/info/:user-id", userHandler.UpdateInfo, userMiddleware.ValidateToken)
	router.PUT("/update/password", userHandler.UpdatePassword, userMiddleware.ValidateToken)
	router.DELETE("/delete/:user-id", userHandler.DeleteUser, userMiddleware.ValidateToken)
}

func (s *server) transactionRouter() {
	router := s.app.Group("/v1/transactions")
	userMiddleware := user_middleware.NewUserMiddleware(s.cfg, s.app.Logger)
	transactionMiddleware := transaction_middleware.NewTransactionMiddleware(s.app.Logger)

	transactionRepository := transaction_repository.NewTransactionRepository(s.db.Connect(), s.app.Logger, s.redisClient)
	transactionService := transaction.NewTransactionService(transactionRepository, s.app.Logger)
	transactionHandler := transaction_handler.NewTransactionHandler(transactionService, s.app.Logger)

	router.GET("/detail/:spender-id", transactionHandler.GetDetails, userMiddleware.ValidateToken, transactionMiddleware.SetGetByTxnTypeRequest)
	router.GET("/summary/:spender-id", transactionHandler.GetSummary, userMiddleware.ValidateToken, transactionMiddleware.SetGetByTxnTypeRequest)
	router.GET("/balance/:spender-id", transactionHandler.GetBalance, userMiddleware.ValidateToken)
	router.GET("/category/:spender-id", transactionHandler.GetByCategory, userMiddleware.ValidateToken, transactionMiddleware.SetGetByCategoryRequest)
	router.GET("/period/:spender-id", transactionHandler.GetByPeriod, userMiddleware.ValidateToken, transactionMiddleware.SetGetByTxnTypeRequest, transactionMiddleware.SetPeriodFilter)
	router.GET("/all", transactionHandler.GetAllTxn, userMiddleware.ValidateToken, transactionMiddleware.SetGetAllTxnFilter, transactionMiddleware.SetTxnPagination)
	router.POST("/save/manual", transactionHandler.SaveByManual, userMiddleware.ValidateToken)
	router.PUT("/update/:txn-id", transactionHandler.Update, userMiddleware.ValidateToken)
	router.DELETE("/delete/:spender-id/:txn-id", transactionHandler.Delete, userMiddleware.ValidateToken)
}
