package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"smart-counter/internal/entity"
	"smart-counter/internal/handlers"
	"smart-counter/internal/middleware"
	"smart-counter/internal/repo"
	"smart-counter/internal/usecase"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const (
	maxDBConnections  = 10
	maxPingDBAttempts = 30
)

func openMySQLConnection() (*sql.DB, error) {
	dsn := "root:"
	mysqlPassword := os.Getenv("pass")
	dsn += mysqlPassword
	dsn += "@tcp(mysql:3306)/golang?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxDBConnections)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	attemptsNumber := 0
	for range ticker.C {
		err = db.Ping()
		attemptsNumber++
		if err == nil {
			break
		}
		if attemptsNumber == maxPingDBAttempts {
			return nil, err
		}
	}
	return db, nil
}

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error in logger start")
		return
	}
	logger := zapLogger.Sugar()
	defer func() {
		err = logger.Sync()
		if err != nil {
			log.Printf("error in logger sync")
		}
	}()
	envFilePath := "./.env"
	err = godotenv.Load(envFilePath)
	if err != nil {
		logger.Fatalf("Error loading .env file: %s", err)
	}
	mySQLDb, err := openMySQLConnection()
	if err != nil {
		logger.Fatalf("error in connection to mysql: %s", err)
	}
	logger.Infof("connected to mysql")
	defer func() {
		err = mySQLDb.Close()
		if err != nil {
			logger.Errorf("error in close connection to mysql: %s", err)
		}
	}()
	computingResources, err := GetComputingResourcesFromDB(mySQLDb)
	if err != nil {
		logger.Fatalf("error in getting computing resources: %s", err)
	}
	operations, err := GetOperations(logger)
	if err != nil {
		logger.Fatalf("error in getting operations: %s", err)
	}
	operationsRepo := repo.NewOperationsMemoryRepo(operations)
	computingResourcesRepo := repo.NewComputingResourceMemoryRepo(mySQLDb, computingResources)
	expressionsRepo := repo.NewExpressionRepoApp(mySQLDb)

	counterUseCase := usecase.NewCounterUseCaseApp(expressionsRepo, operationsRepo, computingResourcesRepo)

	counterHandler := handlers.NewCounterHandler(logger, counterUseCase)

	router := mux.NewRouter()

	router.HandleFunc("/expression", counterHandler.AddExpression).Methods(http.MethodPost)
	router.HandleFunc("/expressions", counterHandler.GetExpressions).Methods(http.MethodGet)
	router.HandleFunc("/operations", counterHandler.GetOperations).Methods(http.MethodGet)
	router.HandleFunc("/operation", counterHandler.SetOperationTime).Methods(http.MethodPut)
	router.HandleFunc("/comp_resources", counterHandler.GetComputingResources).Methods(http.MethodGet)

	router.Use(middleware.RequestInitMiddleware)
	router.Use(middleware.AccessLogMiddleware)

	logger.Infof("starting server on port 8080")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		logger.Fatalf("error in starting server")
	}

}

func GetComputingResourcesFromDB(db *sql.DB) ([]*entity.ComputingResource, error) {
	compResources := []*entity.ComputingResource{}
	rows, err := db.Query("SELECT id, name, status, last_ping_time, parallel_computers_num FROM computing_resource")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			log.Printf("error in closing sql rows: %s\n", err)
		}
	}(rows)
	for rows.Next() {
		compRes := &entity.ComputingResource{}
		err = rows.Scan(&compRes.ID, &compRes.Name, &compRes.Status, &compRes.LastPingTime, &compRes.ParallelComputersNum)
		if err != nil {
			return nil, err
		}
		limitChan := make(chan struct{}, compRes.ParallelComputersNum)
		compRes.LimitChan = limitChan
		compResources = append(compResources, compRes)
	}
	return compResources, nil
}

func GetOperations(logger *zap.SugaredLogger) (map[byte]*entity.Operation, error) {
	opDur := os.Getenv("operations_default_duration")
	opDurInt, err := strconv.Atoi(opDur)
	if err != nil {
		logger.Errorf("error in getting integer of operation duration: %s", err)
		return nil, err
	}
	operationPlus := &entity.Operation{
		Name:          '+',
		ExecutionTime: opDurInt,
		ExecutionFunc: entity.Plus,
	}
	operationMinus := &entity.Operation{
		Name:          '-',
		ExecutionTime: opDurInt,
		ExecutionFunc: entity.Minus,
	}
	operationMultiply := &entity.Operation{
		Name:          '*',
		ExecutionTime: opDurInt,
		ExecutionFunc: entity.Multiply,
	}
	operationDivide := &entity.Operation{
		Name:          '/',
		ExecutionTime: opDurInt,
		ExecutionFunc: entity.Divide,
	}
	operations := make(map[byte]*entity.Operation, 4)
	operations['+'] = operationPlus
	operations['-'] = operationMinus
	operations['*'] = operationMultiply
	operations['/'] = operationDivide
	return operations, nil
}
