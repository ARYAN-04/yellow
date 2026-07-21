package models

import (
	"errors"
	"time"
)

var ErrNotAssigned = errors.New("unauthorized: you are not assigned to judge this debate")

// User represents a global platform user account.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

// Tournament represents a metadata record for a tournament stored in the global DB.
type Tournament struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	DBPath    string    `json:"db_path"`
	CreatedAt time.Time `json:"created_at"`
}

// Participation defines a user's role and permission scope within a tournament.
type Participation struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	TournamentID string `json:"tournament_id"`
	Role         string `json:"role"` // 'organizer', 'adjudicator', 'team'
}

// Institution represents a tournament institution.
type Institution struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Speaker represents a team speaker.
type Speaker struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Team represents a tournament team with its speakers and institution.
type Team struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	InstitutionID   *string   `json:"institution_id"`
	InstitutionName *string   `json:"institution_name"`
	InstitutionCode *string   `json:"institution_code"`
	Speakers        []Speaker `json:"speakers"`
}

// Adjudicator represents a tournament adjudicator.
type Adjudicator struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	InstitutionID   *string `json:"institution_id"`
	InstitutionName *string `json:"institution_name"`
	InstitutionCode *string `json:"institution_code"`
	TestScore       float64 `json:"test_score"`
}

// Round represents a tournament round.
type Round struct {
	ID              string `json:"id"`
	Seq             int    `json:"seq"`
	Name            string `json:"name"`
	Stage           string `json:"stage"` // 'preliminary', 'elimination'
	Silent          bool   `json:"silent"`
	DrawReleased    bool   `json:"draw_released"`
	ResultsReleased bool   `json:"results_released"`
}

// TeamAssignment represents a team assigned to a side in a debate.
type TeamAssignment struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	Side     string `json:"side"`
}

// AdjudicatorAssignment represents an adjudicator assigned to a role in a debate.
type AdjudicatorAssignment struct {
	AdjudicatorID   string `json:"adjudicator_id"`
	AdjudicatorName string `json:"adjudicator_name"`
	Role            string `json:"role"` // 'chair', 'panel', 'trainee'
}

// DebateDraw represents a debate round pairing.
type DebateDraw struct {
	ID           string                  `json:"id"`
	Venue        string                  `json:"venue"`
	Teams        []TeamAssignment        `json:"teams"`
	Adjudicators []AdjudicatorAssignment `json:"adjudicators"`
}

// TeamBallotResult represents the result points for a team.
type TeamBallotResult struct {
	TeamID        string  `json:"team_id"`
	Points        int     `json:"points"`
	SpeakerPoints float64 `json:"speaker_points"`
}

// Standing represents a team's current rank and scores.
type Standing struct {
	TeamID          string  `json:"team_id"`
	TeamName        string  `json:"team_name"`
	TeamCode        string  `json:"team_code"`
	InstitutionCode string  `json:"institution_code"`
	Points          int     `json:"points"`
	SpeakerPoints   float64 `json:"speaker_points"`
}

// TokenInfo represents the resolved access token details.
type TokenInfo struct {
	Slug      string `json:"slug"`
	Type      string `json:"type"` // 'team', 'adjudicator'
	OwnerID   string `json:"owner_id"`
	OwnerName string `json:"owner_name"`
}

// DebateInfo represents a debate for a participant portal.
type DebateInfo struct {
	ID            string           `json:"id"`
	Venue         string           `json:"venue"`
	RoundName     string           `json:"round_name"`
	RoundSeq      int              `json:"round_seq"`
	Role          string           `json:"role,omitempty"` // 'chair', 'panel'
	Side          string           `json:"side,omitempty"`
	Teams         []TeamAssignment `json:"teams"`
	BallotStatus  string           `json:"ballot_status,omitempty"`
	Points        *int             `json:"points,omitempty"`
	SpeakerPoints *float64         `json:"speaker_points,omitempty"`
}

// TokenOwner represents the resolved token owner type and ID.
type TokenOwner struct {
	Type    string
	OwnerID string
}

// SpeakerRequest represents the payload for speaker creation.
type SpeakerRequest struct {
	Name string `json:"name"`
}

// TeamImport holds team import fields parsed from CSV.
type TeamImport struct {
	Name            string
	Code            string
	InstitutionCode string
	Speakers        []string
}

// AdjudicatorImport holds adjudicator import fields parsed from CSV.
type AdjudicatorImport struct {
	Name            string
	InstitutionCode string
	TestScore       float64
}

// Draw-specific models to communicate between draw algorithm and store:

type TeamDrawInfo struct {
	ID            string
	InstitutionID string
}

type SideHistKey struct {
	TeamID string
	Side   string
}

type AdjDrawInfo struct {
	ID            string
	InstitutionID string
	Score         float64
}

type DebateDrawInput struct {
	DebateID     string
	Venue        string
	Teams        []TeamAssignment
	Adjudicators []AdjudicatorAssignment
}
