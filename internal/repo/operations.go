package repo

import (
	"smart-counter/internal/entity"
	errorapp "smart-counter/internal/errors"
)

type OperationsRepo interface {
	GetOperations() ([]*entity.Operation, error)
	SetOperationTime(operation *entity.Operation) (*entity.Operation, error)
	GetOperation(name byte) (*entity.Operation, error)
}

type OperationsMemoryRepo struct {
	operations map[byte]*entity.Operation
}

func NewOperationsMemoryRepo(operations map[byte]*entity.Operation) *OperationsMemoryRepo {
	return &OperationsMemoryRepo{
		operations: operations,
	}
}

func (or *OperationsMemoryRepo) GetOperations() ([]*entity.Operation, error) {
	operations := make([]*entity.Operation, 0)
	for _, op := range or.operations {
		operations = append(operations, op)
	}
	return operations, nil
}

func (or *OperationsMemoryRepo) SetOperationTime(operation *entity.Operation) (*entity.Operation, error) {
	operationToChange, exists := or.operations[operation.Name]
	if !exists {
		return nil, errorapp.ErrorUnknownOperation
	}
	operationToChange.ExecutionTime = operation.ExecutionTime
	return operationToChange, nil
}

func (or *OperationsMemoryRepo) GetOperation(name byte) (*entity.Operation, error) {
	op, exists := or.operations[name]
	if !exists {
		return nil, errorapp.ErrorUnknownOperation
	}
	return op, nil
}
