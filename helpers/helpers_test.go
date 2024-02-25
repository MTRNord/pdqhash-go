package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTorben(t *testing.T) {
	numRows := 4
	numCols := 8
	m := make([][]float64, numRows)
	for i := range m {
		m[i] = make([]float64, numCols)
		for j := 0; j < numCols; j++ {
			m[i][j] = float64(i) + (float64(j) * 0.01)
		}
	}

	assert.Equal(t, float64(1.07), Torben(m, numRows, numCols), "The Torben function should produce 1.07 for 3 rows and 8 cols")
}

func TestAbs(t *testing.T) {
	assert.Equal(t, int(1), Abs(-1), "The absolute value of -1 should be 1")
	assert.Equal(t, int(1), Abs(1), "The absolute value of 1 should be 1")
	assert.Equal(t, int(0), Abs(0), "The absolute value of 0 should be 0")
}
