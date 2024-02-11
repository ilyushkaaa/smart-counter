package entity

type ComputingResource struct {
	ID                   int
	Name                 string
	Status               string
	LastPingTime         string
	LimitChan            chan struct{}
	ParallelComputersNum int
}
