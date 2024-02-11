package middleware

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"net/http"
	errorapp "smart-counter/internal/errors"
)

type loggerKey int

const MyLoggerKey loggerKey = 1

func RequestInitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		myLogger, err := initLogger()
		if err != nil {
			WriteNoLoggerResponse(w)
		}
		requestID := uuid.New().String()
		myLogger = myLogger.With(zap.String("request-id", requestID))
		ctx := r.Context()
		ctx = context.WithValue(ctx, MyLoggerKey, myLogger)
		myLogger.Infof("request init middleware call")
		next.ServeHTTP(w, r.WithContext(ctx))
		loggerErr := myLogger.Sync() // Вызываем Sync() в конце обработки
		if loggerErr != nil {
			log.Println("error in logger sync")
		}
	})
}

func initLogger() (*zap.SugaredLogger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error in logger initialization: %s", err)
		return nil, err
	}
	myLogger := zapLogger.Sugar()
	return myLogger, nil
}

func GetLoggerFromContext(ctx context.Context) (*zap.SugaredLogger, error) {
	myLogger, ok := ctx.Value(MyLoggerKey).(*zap.SugaredLogger)
	if !ok {
		return myLogger, errorapp.ErrorNoLogger
	}
	return myLogger, nil
}

func WriteNoLoggerResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(`{"message":"internal error"}`))
	if err != nil {
		log.Printf("error in writing response body: %s", err)
	}
}
