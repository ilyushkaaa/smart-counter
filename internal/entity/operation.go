package entity

type Operation struct {
	ExecutionFunc func(num1, num2 float64) float64
	ExecutionTime int
	Name          byte
}

func Plus(num1, num2 float64) float64 {
	return num1 + num2
}

func Minus(num1, num2 float64) float64 {
	return num1 - num2
}

func Multiply(num1, num2 float64) float64 {
	return num1 * num2
}

func Divide(num1, num2 float64) float64 {
	return num1 / num2
}
