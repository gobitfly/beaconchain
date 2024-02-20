package apitypes

type Sort[T ~int] struct {
	Column T
	IsDesc bool
}
