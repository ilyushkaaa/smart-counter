package dto

import "database/sql"

type ExpressionResult struct {
	FinishTime   string
	Result       sql.NullFloat64
	IsSuccessful bool
}

type ExpressionSQL struct {
	ID           string
	Body         string
	Result       sql.NullFloat64
	CreationTime string
	FinishTime   sql.NullString
	IsFinished   bool
	IsSuccessful sql.NullBool
}

type OperationChange struct {
	ExecutionTime string
	Name          byte
}

type NumInExpression struct {
	Value      float64
	StartIndex int
	EndIndex   int
}

type OperationInExpression struct {
	LeftNumIndex int
	Value        byte
	WasUsed      bool
}

type ExpressionToAdd struct {
	Body string
}
