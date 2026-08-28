package draw

import (
	"fmt"
	"yellow/internal/models"
)

// AllocateVenues assigns available venues to debates in order of bracket importance.
// If not enough venues exist, it falls back to "Room N" naming.
func AllocateVenues(order []int, debates []models.DebateDrawInput, venues []models.Venue) {
	for rank, debateIdx := range order {
		if rank < len(venues) {
			debates[debateIdx].Venue = venues[rank].Name
		} else {
			debates[debateIdx].Venue = fmt.Sprintf("Room %d", rank+1)
		}
	}
}
