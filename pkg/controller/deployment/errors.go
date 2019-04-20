package deployment

import "errors"

var (
	ErrObjectChanged = errors.New("object changed and can't be updated")
)
