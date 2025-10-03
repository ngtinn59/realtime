package errs

import (
	"errors"
	"fmt"
)

type OrderNotExistsError struct {
	OrderID int64
}
type ErrProductNotExists struct {
	ProductID int64
}
type ErrProductExists struct {
	ProductID int64
}

func (e *OrderNotExistsError) Error() string {
	return fmt.Sprintf("order with ID %d does not exist", e.OrderID)
}
func (e *ErrProductNotExists) Error() string {
	return fmt.Sprintf("product with ID %d does not exist", e.ProductID)
}

func (e *ErrProductExists) Error() string {
	return fmt.Sprintf("product with ID %d already exists", e.ProductID)
}

var (
	ErrMaterialExists = errors.New("material already exists")
)
