package repo

import (
	"database/sql"
	"errors"
	"log"
	"smart-counter/internal/dto"
	"smart-counter/internal/entity"

	_ "github.com/go-sql-driver/mysql"
)

type ExpressionRepo interface {
	AddExpression(expression *entity.Expression) (*entity.Expression, error)
	GetExpressions() ([]*entity.Expression, error)
	AddExpressionResult(exprResult dto.ExpressionResult, ID string) error
}

type ExpressionRepoApp struct {
	db *sql.DB
}

func NewExpressionRepoApp(db *sql.DB) *ExpressionRepoApp {
	return &ExpressionRepoApp{
		db: db,
	}
}

func (er *ExpressionRepoApp) AddExpression(expression *entity.Expression) (*entity.Expression, error) {
	_, err := er.db.Exec(
		"INSERT INTO expressions (`id`, `body`, `creation_time`, `is_finished`) VALUES (?, ?, ?, ?, ?)",
		expression.ID,
		expression.Body,
		expression.CreationTime,
		expression.IsFinished,
		expression.IsSuccessFul,
	)
	if err != nil {
		return nil, err
	}
	return expression, nil
}

func (er *ExpressionRepoApp) GetExpressions() ([]*entity.Expression, error) {
	expressions := []*entity.Expression{}
	rows, err := er.db.Query("SELECT id, body, result, creation_time, finish_time, is_finished, is_successful FROM expressions")
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
		expressionSQL := &dto.ExpressionSQL{}
		err = rows.Scan(&expressionSQL.ID, &expressionSQL.Body, &expressionSQL.Result, &expressionSQL.CreationTime, &expressionSQL.FinishTime, &expressionSQL.IsFinished, &expressionSQL.IsSuccessful)
		if err != nil {
			return nil, err
		}
		expression := &entity.Expression{
			ID:           expressionSQL.ID,
			Body:         expressionSQL.Body,
			CreationTime: expressionSQL.CreationTime,
			IsFinished:   expressionSQL.IsFinished,
		}
		if expressionSQL.FinishTime.Valid {
			expression.FinishTime = expressionSQL.FinishTime.String
		}
		if expressionSQL.Result.Valid {
			expression.Result = expressionSQL.Result.Float64
		}
		if expressionSQL.IsSuccessful.Valid {
			expression.IsSuccessFul = expressionSQL.IsSuccessful.Bool
		}
		expressions = append(expressions, expression)
	}
	return expressions, nil
}

func (er *ExpressionRepoApp) AddExpressionResult(exprResult dto.ExpressionResult, ID string) error {
	_, err := er.db.Exec(
		"UPDATE expressions SET is_successful = ?, result = ?, finish_time = ? where id = ?",
		exprResult.IsSuccessful,
		exprResult.Result,
		exprResult.FinishTime,
		ID,
	)
	return err
}
