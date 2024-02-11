package entity

type Expression struct {
	ID           string
	Body         string
	Result       float64
	CreationTime string
	FinishTime   string
	IsFinished   bool
	IsSuccessFul bool
}
