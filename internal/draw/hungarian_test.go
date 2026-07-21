package draw

import (
	"testing"
)

func TestSolveHungarian(t *testing.T) {
	matrix := [][]int{
		{10, 19, 8, 15},
		{10, 18, 7, 17},
		{13, 16, 9, 14},
		{12, 17, 8, 16},
	}

	assignment := SolveHungarian(matrix)

	// Verify the size of the result matching N
	if len(assignment) != 4 {
		t.Fatalf("expected assignment length of 4, got %d", len(assignment))
	}

	// Compute total cost of the returned assignment
	totalCost := 0
	for i, j := range assignment {
		totalCost += matrix[i][j]
	}

	// The minimum cost assignment is:
	// Row 0 -> Col 0 (10)
	// Row 1 -> Col 2 (7)
	// Row 2 -> Col 3 (14)
	// Row 3 -> Col 1 (17)
	// Total cost = 48
	if totalCost != 48 {
		t.Errorf("expected minimum cost of 48, got %d (assignment: %v)", totalCost, assignment)
	}
}
