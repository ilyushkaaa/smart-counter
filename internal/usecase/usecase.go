package usecase

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"smart-counter/internal/dto"
	"smart-counter/internal/entity"
	errorapp "smart-counter/internal/errors"
	"smart-counter/internal/repo"
	"strconv"
	"sync"
	"time"
)

type CounterUseCase interface {
	AddExpression(expression string, logger *zap.SugaredLogger) (*entity.Expression, error)
	GetExpressions(logger *zap.SugaredLogger) ([]*entity.Expression, error)
	GetComputingResources(logger *zap.SugaredLogger) ([]*entity.ComputingResource, error)
	GetOperations(logger *zap.SugaredLogger) ([]*entity.Operation, error)
	SetOperationTime(operation dto.OperationChange, logger *zap.SugaredLogger) (*entity.Operation, error)
}

type CounterUseCaseApp struct {
	expressionsRepo       repo.ExpressionRepo
	operationsRepo        repo.OperationsRepo
	computingResourceRepo repo.ComputingResourceRepo
}

func NewCounterUseCaseApp(expressionsRepo repo.ExpressionRepo, operationsRepo repo.OperationsRepo,
	computingResourceRepo repo.ComputingResourceRepo) *CounterUseCaseApp {
	return &CounterUseCaseApp{
		expressionsRepo:       expressionsRepo,
		operationsRepo:        operationsRepo,
		computingResourceRepo: computingResourceRepo,
	}
}

func (cu *CounterUseCaseApp) AddExpression(expression string, logger *zap.SugaredLogger) (*entity.Expression, error) {
	nums, operations, countPriorityOps, err := cu.splitNumsAndOps(expression, logger)
	if err != nil {
		return nil, err
	}
	logger.Infof("nums: %v", nums)
	logger.Infof("operations: %v", operations)
	exprID := uuid.New().String()
	newExpression := &entity.Expression{
		ID:           exprID,
		Body:         expression,
		CreationTime: time.Now().String(),
		IsFinished:   false,
	}
	addedExpression, err := cu.expressionsRepo.AddExpression(newExpression)
	if err != nil {
		logger.Errorf("error in adding expression to repository: %s", err)
		return nil, err
	}
	computingResource, err := cu.computingResourceRepo.GetComputingResourceForTask()
	if err != nil {
		logger.Errorf("no available computing resources")
		return nil, err
	}
	computingResource, err = cu.computingResourceRepo.ChangeComputingResource(computingResource, "ok", time.Now().String())
	if err != nil {
		logger.Errorf("error in changing last ping time for computing resource: %s", err)
		return nil, err
	}
	go cu.computeExpression(computingResource, nums, operations, logger, countPriorityOps, exprID)
	return addedExpression, nil
}

func (cu *CounterUseCaseApp) GetExpressions(logger *zap.SugaredLogger) ([]*entity.Expression, error) {
	exprs, err := cu.expressionsRepo.GetExpressions()
	if err != nil {
		logger.Errorf("error in getting expressions: %s\n", err)
		return nil, err
	}
	return exprs, nil
}

func (cu *CounterUseCaseApp) GetComputingResources(logger *zap.SugaredLogger) ([]*entity.ComputingResource, error) {
	compRes, err := cu.computingResourceRepo.GetComputingResources()
	if err != nil {
		logger.Errorf("error in getting computing resources: %s\n", err)
	}
	return compRes, nil
}

func (cu *CounterUseCaseApp) GetOperations(logger *zap.SugaredLogger) ([]*entity.Operation, error) {
	operations, err := cu.operationsRepo.GetOperations()
	if err != nil {
		logger.Errorf("error in getting opeartions: %s\n", err)
	}
	return operations, nil
}

func (cu *CounterUseCaseApp) SetOperationTime(operationDTO dto.OperationChange, logger *zap.SugaredLogger) (*entity.Operation, error) {
	newExecTime, err := strconv.Atoi(operationDTO.ExecutionTime)
	if err != nil {
		logger.Errorf("can not cast execution time to integer: %s\n", err)
		return nil, errorapp.ErrorBadExecutionTime
	}
	operation := &entity.Operation{
		Name:          operationDTO.Name,
		ExecutionTime: newExecTime,
	}
	changedOperation, err := cu.operationsRepo.SetOperationTime(operation)
	if err != nil {
		logger.Errorf("error in changing execution time for operation: %s\n", err)
		return nil, err
	}
	return changedOperation, nil
}

func (cu *CounterUseCaseApp) splitNumsAndOps(expression string, logger *zap.SugaredLogger) ([]dto.NumInExpression, []dto.OperationInExpression, int, error) {
	nums := make([]dto.NumInExpression, 0)
	operations := make([]dto.OperationInExpression, 0)
	currentStr := make([]byte, 0)
	isLastOp := false
	numCounter := 0
	countPriorityOps := 0
	for i := 0; i < len(expression); i++ {
		if expression[i] >= 48 && expression[i] <= 57 {
			if len(currentStr) == 0 {
				isLastOp = false
			} else {
				if currentStr[0] == '0' {
					logger.Errorf("num can not start from 0")
					return nil, nil, 0, errorapp.ErrorInvalidInput
				}
			}
			currentStr = append(currentStr, expression[i])
		} else {
			if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
				if isLastOp {
					logger.Errorf("two operations can not be neighbours")
					return nil, nil, 0, errorapp.ErrorInvalidInput
				}
				if expression[i] == '*' || expression[i] == '/' {
					countPriorityOps++
				}
				num := string(currentStr)
				numFloat, err := strconv.ParseFloat(num, 64)
				if err != nil {
					logger.Errorf("wrong format of num")
					return nil, nil, 0, errorapp.ErrorInvalidInput
				}
				newNum := dto.NumInExpression{
					Value:      numFloat,
					StartIndex: numCounter,
					EndIndex:   numCounter,
				}
				nums = append(nums, newNum)
				isLastOp = true
				newOperation := dto.OperationInExpression{
					Value:        expression[i],
					WasUsed:      false,
					LeftNumIndex: numCounter,
				}
				numCounter++
				operations = append(operations, newOperation)
				currentStr = make([]byte, 0)

			} else {
				logger.Errorf("expression contains symbol which is not allowed")
				return nil, nil, 0, errorapp.ErrorInvalidInput
			}
		}
	}
	if len(currentStr) != 0 {
		num := string(currentStr)
		numFloat, err := strconv.ParseFloat(num, 64)
		if err != nil {
			logger.Errorf("wrong format of num")
			return nil, nil, 0, errorapp.ErrorInvalidInput
		}
		newNum := dto.NumInExpression{
			Value:      numFloat,
			StartIndex: numCounter,
			EndIndex:   numCounter,
		}
		nums = append(nums, newNum)
	} else {
		logger.Errorf("empty input or operation in the end of expression")
		return nil, nil, 0, errorapp.ErrorInvalidInput
	}
	if len(operations) == 0 {
		logger.Errorf("at least one operation required")
		return nil, nil, 0, errorapp.ErrorInvalidInput
	}
	return nums, operations, countPriorityOps, nil
}

func (cu *CounterUseCaseApp) computeExpression(compResource *entity.ComputingResource, nums []dto.NumInExpression, operations []dto.OperationInExpression, logger *zap.SugaredLogger, countPriorityOps int, exprID string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opsLeft := countPriorityOps
	lastUsedIndex := 0
	for countPriorityOps > 0 {
		diffBetweenOps := 1
		wg := &sync.WaitGroup{}
		for i := 0; i < len(operations); i++ {
			if !operations[i].WasUsed {
				diffBetweenOps++
				if operations[i].Value == '*' || operations[i].Value == '/' {
					if diffBetweenOps > 1 {
						countPriorityOps--
						diffBetweenOps = 1
						wg.Add(1)
						operations[i].WasUsed = true
						if i > 0 && operations[i-1].Value == '/' {
							if operations[i].Value == '/' {
								operations[i].Value = '*'
							} else {
								operations[i].Value = '/'
							}
						}
						lastUsedIndex = i
						currentOperation, err := cu.operationsRepo.GetOperation(operations[i].Value)
						if err != nil {
							logger.Errorf("error in getting operation")
							exprResult := dto.ExpressionResult{
								FinishTime:   time.Now().String(),
								IsSuccessful: false,
							}
							err = cu.expressionsRepo.AddExpressionResult(exprResult, exprID)
							if err != nil {
								logger.Errorf("error in adding expression result to repository: %s", err)
							}
							return
						}
						go cu.performOperation(cancel, wg, currentOperation, compResource, nums, operations, i)

					}
				}

			}
		}
		waitChan := make(chan struct{})
		go waitFunc(wg, waitChan)
		select {
		case <-ctx.Done():
			return
		case <-waitChan:
			continue
		}

	}

	countNonPriorityOps := len(operations) - opsLeft
	for countNonPriorityOps > 0 {
		diffBetweenOps := 1
		wg := &sync.WaitGroup{}
		for i := 0; i < len(operations); i++ {
			if !operations[i].WasUsed {
				diffBetweenOps++
				if operations[i].Value == '+' || operations[i].Value == '-' {
					if diffBetweenOps > 1 {
						countPriorityOps--
						diffBetweenOps = 1
						wg.Add(1)
						operations[i].WasUsed = true
						if i > 0 && operations[i-1].Value == '-' {
							if operations[i].Value == '-' {
								operations[i].Value = '+'
							} else {
								operations[i].Value = '-'
							}
						}
						lastUsedIndex = i
						currentOperation, err := cu.operationsRepo.GetOperation(operations[i].Value)
						if err != nil {
							logger.Errorf("error in getting operation")
							exprResult := dto.ExpressionResult{
								FinishTime:   time.Now().String(),
								IsSuccessful: false,
							}
							err = cu.expressionsRepo.AddExpressionResult(exprResult, exprID)
							if err != nil {
								logger.Errorf("error in adding expression result to repository: %s", err)
							}
							return
						}
						go cu.performOperation(cancel, wg, currentOperation, compResource, nums, operations, i)

					}
				}

			}
		}
		waitChan := make(chan struct{})
		go waitFunc(wg, waitChan)
		select {
		case <-ctx.Done():
			logger.Errorf("request timeout")
			return
		case <-waitChan:
			continue
		}

	}
	finalResult := nums[nums[operations[lastUsedIndex].LeftNumIndex].StartIndex].Value
	exprRes := dto.ExpressionResult{
		Result: sql.NullFloat64{
			Float64: finalResult,
			Valid:   true,
		},
		FinishTime:   time.Now().String(),
		IsSuccessful: true,
	}
	err := cu.expressionsRepo.AddExpressionResult(exprRes, exprID)
	if err != nil {
		logger.Errorf("error in adding successful result of expression computing in repository: %s", err)
	}

}

func waitFunc(wg *sync.WaitGroup, waitChan chan struct{}) {
	wg.Wait()
	waitChan <- struct{}{}
}

func (cu *CounterUseCaseApp) performOperation(cancel context.CancelFunc, wg *sync.WaitGroup, currentOperation *entity.Operation, compResource *entity.ComputingResource, nums []dto.NumInExpression, operations []dto.OperationInExpression, currentOpIndex int) {
	deadline := time.NewTimer(time.Second * 5)
	resChan := make(chan float64)
	go func() {
		compResource.LimitChan <- struct{}{}
		defer func() {
			<-compResource.LimitChan
		}()
		time.Sleep(time.Second * time.Duration(currentOperation.ExecutionTime))
		resChan <- currentOperation.ExecutionFunc(nums[operations[currentOpIndex].LeftNumIndex].Value, nums[operations[currentOpIndex].LeftNumIndex+2].Value)
	}()
	select {
	case <-deadline.C:
		_, err := cu.computingResourceRepo.ChangeComputingResource(compResource, "restart", time.Now().String())
		if err != nil {
			cancel()
			return
		}
	case res := <-resChan:
		nums[nums[operations[currentOpIndex].LeftNumIndex].StartIndex].Value = res
		nums[nums[operations[currentOpIndex].LeftNumIndex+2].EndIndex].Value = res
		wg.Done()
		return
	}

}
