package utils

import (
	"github.com/libopenstorage/cloudops"
)

// CopyDecisionMatrix creates a copy of the decision matrix
func CopyDecisionMatrix(matrix *cloudops.StorageDecisionMatrix) *cloudops.StorageDecisionMatrix {
	matrixCopy := &cloudops.StorageDecisionMatrix{}
	matrixCopy.Rows = make([]cloudops.StorageDecisionMatrixRow, len(matrix.Rows))
	copy(matrixCopy.Rows, matrix.Rows)
	return matrixCopy
}
