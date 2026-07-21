package draw

// SolveHungarian implements the Kuhn-Munkres (Hungarian) algorithm for minimum cost bipartite matching.
// It takes a square matrix of size N x N representing the costs and returns an array of size N
// where result[i] is the column index assigned to row i.
func SolveHungarian(matrix [][]int) []int {
	n := len(matrix)
	if n == 0 {
		return nil
	}

	// Create a local copy of the cost matrix to preserve the input
	cost := make([][]int, n)
	for i := range matrix {
		cost[i] = make([]int, n)
		copy(cost[i], matrix[i])
	}

	// u, v are potentials; p tracks matching columns; way stores the alternating path
	u := make([]int, n+1)
	v := make([]int, n+1)
	p := make([]int, n+1)
	way := make([]int, n+1)

	// Inf represents infinity for cost initialization
	const Inf = 1000000000

	for i := 1; i <= n; i++ {
		p[0] = i
		j0 := 0
		minv := make([]int, n+1)
		for j := range minv {
			minv[j] = Inf
		}
		used := make([]bool, n+1)

		for {
			used[j0] = true
			i0 := p[j0]
			delta := Inf
			j1 := 0

			for j := 1; j <= n; j++ {
				if !used[j] {
					cur := cost[i0-1][j-1] - u[i0] - v[j]
					if cur < minv[j] {
						minv[j] = cur
						way[j] = j0
					}
					if minv[j] < delta {
						delta = minv[j]
						j1 = j
					}
				}
			}

			for j := 0; j <= n; j++ {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}

			j0 = j1
			if p[j0] == 0 {
				break
			}
		}

		for {
			j1 := way[j0]
			p[j0] = p[j1]
			j0 = j1
			if j0 == 0 {
				break
			}
		}
	}

	assignment := make([]int, n)
	for j := 1; j <= n; j++ {
		assignment[p[j]-1] = j - 1
	}

	return assignment
}
