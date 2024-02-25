package helpers

import "golang.org/x/exp/constraints"

func Abs[T constraints.Integer](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func Torben(m [][]float64, numRows, numCols int) float64 {
	n := numRows * numCols
	midn := (n + 1) / 2
	less := 0
	greater := 0
	equal := 0
	min := m[0][0]
	max := m[0][0]
	guess := float64(0.0)
	maxltguess := float64(0.0)
	mingtguess := float64(0.0)
	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			v := m[i][j]
			if v < min {
				min = v
			}

			if v > max {
				max = v
			}
		}
	}

	for {
		guess = (min + max) / 2
		less = 0
		greater = 0
		equal = 0
		maxltguess = min
		mingtguess = max

		for i := 0; i < numRows; i++ {
			for j := 0; j < numCols; j++ {
				v := m[i][j]
				if v < guess {
					less++
					if v > maxltguess {
						maxltguess = v
					}
				} else if v > guess {
					greater++
					if v < mingtguess {
						mingtguess = v
					}
				} else {
					equal++
				}
			}
		}

		if less <= midn && greater <= midn {
			break
		} else if less > greater {
			max = maxltguess
		} else {
			min = mingtguess
		}
	}

	if less >= midn {
		return maxltguess
	} else if less+equal >= midn {
		return guess
	} else {
		return mingtguess
	}
}
