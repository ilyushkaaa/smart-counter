package errorapp

import "errors"

var (
	ErrorNoLogger             = errors.New("no logger in context")
	ErrorUnknownOperation     = errors.New("unknown operation")
	ErrorNoComputingResources = errors.New("no active available computing resources")
	ErrorBadExecutionTime     = errors.New("execution time mast be integer number of seconds")
	ErrorInvalidInput         = errors.New("expression failed validation")
)
