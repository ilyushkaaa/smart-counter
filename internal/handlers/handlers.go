package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"smart-counter/internal/dto"
	errorapp "smart-counter/internal/errors"
	"smart-counter/internal/middleware"
	"smart-counter/internal/usecase"
)

type CounterHandler struct {
	logger   *zap.SugaredLogger
	useCases usecase.CounterUseCase
}

func NewCounterHandler(logger *zap.SugaredLogger, useCases usecase.CounterUseCase) *CounterHandler {
	return &CounterHandler{
		logger:   logger,
		useCases: useCases,
	}
}

func (ch *CounterHandler) AddExpression(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := middleware.GetLoggerFromContext(ctx)
	if err != nil {
		log.Printf("can not get logger from context: %s", err)
		middleware.WriteNoLoggerResponse(w)
	}
	rBody, err := io.ReadAll(r.Body)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			logger.Errorf("error in closing request body: %s", err)
		}
	}()
	expressionToAdd := &dto.ExpressionToAdd{}
	err = json.Unmarshal(rBody, expressionToAdd)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in decoding expression: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	expression, err := ch.useCases.AddExpression(expressionToAdd.Body, logger)
	if errors.Is(err, errorapp.ErrorInvalidInput) || errors.Is(err, errorapp.ErrorUnknownOperation) {
		errText := fmt.Sprintf(`{"message": "bad expression format: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	if errors.Is(err, errorapp.ErrorNoComputingResources) {
		errText := fmt.Sprintf(`{"message": "error with computing resoyrces: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	if err != nil {
		errText := `{"message": "internal server error"}`
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	reviewJSON, err := json.Marshal(expression)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding expression: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	WriteResponse(logger, w, reviewJSON, http.StatusOK)

}

func (ch *CounterHandler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := middleware.GetLoggerFromContext(ctx)
	if err != nil {
		log.Printf("can not get logger from context: %s", err)
		middleware.WriteNoLoggerResponse(w)
	}
	expressions, err := ch.useCases.GetExpressions(logger)
	if err != nil {
		errText := `{"message": "internal server error"}`
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	expressionsJSON, err := json.Marshal(expressions)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding expressions: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	WriteResponse(logger, w, expressionsJSON, http.StatusOK)

}

func (ch *CounterHandler) GetOperations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := middleware.GetLoggerFromContext(ctx)
	if err != nil {
		log.Printf("can not get logger from context: %s", err)
		middleware.WriteNoLoggerResponse(w)
	}
	operations, err := ch.useCases.GetOperations(logger)
	if err != nil {
		errText := `{"message": "internal server error"}`
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	operationsJSON, err := json.Marshal(operations)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding operations: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	WriteResponse(logger, w, operationsJSON, http.StatusOK)
}

func (ch *CounterHandler) GetComputingResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := middleware.GetLoggerFromContext(ctx)
	if err != nil {
		log.Printf("can not get logger from context: %s", err)
		middleware.WriteNoLoggerResponse(w)
	}
	compRes, err := ch.useCases.GetComputingResources(logger)
	if err != nil {
		errText := `{"message": "internal server error"}`
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	compResJSON, err := json.Marshal(compRes)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding computing resources: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	WriteResponse(logger, w, compResJSON, http.StatusOK)
}

func (ch *CounterHandler) SetOperationTime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := middleware.GetLoggerFromContext(ctx)
	if err != nil {
		log.Printf("can not get logger from context: %s", err)
		middleware.WriteNoLoggerResponse(w)
	}
	rBody, err := io.ReadAll(r.Body)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			logger.Errorf("error in closing request body: %s", err)
		}
	}()
	operationToChange := dto.OperationChange{}
	err = json.Unmarshal(rBody, &operationToChange)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in decoding operation: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	expression, err := ch.useCases.SetOperationTime(operationToChange, logger)
	if errors.Is(err, errorapp.ErrorUnknownOperation) {
		errText := fmt.Sprintf(`{"message": "unknown operation: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	if errors.Is(err, errorapp.ErrorBadExecutionTime) {
		errText := fmt.Sprintf(`{"message": "execition time must be integer num of seconds: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	if err != nil {
		errText := `{"message": "internal server error"}`
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	operationJSON, err := json.Marshal(expression)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding operation: %s"}`, err)
		WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	WriteResponse(logger, w, operationJSON, http.StatusOK)
}

func WriteResponse(logger *zap.SugaredLogger, w http.ResponseWriter, dataJSON []byte, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	_, err := w.Write(dataJSON)
	if err != nil {
		logger.Errorf("error in writing response body: %s", err)
	}
}
