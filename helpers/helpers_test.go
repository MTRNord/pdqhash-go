package helpers

import "testing"

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

	if Torben(m, numRows, numCols) != 1.07 {
		t.Errorf("Incorrect median")
	}
}
