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
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsNovice bool   `json:"is_novice"`
	IsEsl    bool   `json:"is_esl"`
	IsEfl    bool   `json:"is_efl"`
}

// Team represents a tournament team with its speakers and institution.
type Team struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	InstitutionID   *string   `json:"institution_id"`
	InstitutionName *string   `json:"institution_name"`
	InstitutionCode *string   `json:"institution_code"`
	IsNovice        bool      `json:"is_novice"`
	IsEsl           bool      `json:"is_esl"`
	IsEfl           bool      `json:"is_efl"`
	IsStandby       bool      `json:"is_standby"`
	Speakers        []Speaker `json:"speakers"`
}

// BreakCategory represents a break qualification category.
type BreakCategory struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Seq        int    `json:"seq"`
	Size       *int   `json:"size"`
	BasePoints *int   `json:"base_points"`
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
	ID       string `json:"id,omitempty"`
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	Side     string `json:"side"`
	PullUp   bool   `json:"pull_up,omitempty"`
}

// AdjudicatorAssignment represents an adjudicator assigned to a role in a debate.
type AdjudicatorAssignment struct {
	ID              string `json:"id,omitempty"`
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
// AdjudicatorID is set on split (per-judge) ballots; NULL means consensus.
type TeamBallotResult struct {
	TeamID        string  `json:"team_id"`
	Points        int     `json:"points"`
	SpeakerPoints float64 `json:"speaker_points"`
	AdjudicatorID *string `json:"adjudicator_id,omitempty"`
}

// BallotSummary represents a ballot and its results, used by the registry view.
type BallotSummary struct {
	ID            string             `json:"id"`
	DebateID      string             `json:"debate_id"`
	DebateVenue   string             `json:"debate_venue"`
	SubmitterType string             `json:"submitter_type"`
	SubmitterName string             `json:"submitter_name"`
	Status        string             `json:"status"`
	IsSplit       bool               `json:"is_split"`
	EntryGroup    *string            `json:"entry_group,omitempty"`
	Results       []TeamBallotResult `json:"results"`
}

// BallotDiff describes a single field mismatch between two double-entry ballots.
type BallotDiff struct {
	TeamID        string   `json:"team_id"`
	AdjudicatorID *string  `json:"adjudicator_id,omitempty"`
	Field         string   `json:"field"` // 'points', 'speaker_points'
	BallotA       *float64 `json:"ballot_a"`
	BallotB       *float64 `json:"ballot_b"`
}

// Standing represents a team's current rank and scores.
type Standing struct {
	TeamID          string  `json:"team_id"`
	TeamName        string  `json:"team_name"`
	TeamCode        string  `json:"team_code"`
	InstitutionCode string  `json:"institution_code"`
	Points          int     `json:"points"`
	SpeakerPoints   float64 `json:"speaker_points"`
	Margin          float64 `json:"margin"`
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

// BreakTeam represents one team's position in a break category.
type BreakTeam struct {
	Rank          int     `json:"rank"`
	TeamID        string  `json:"team_id"`
	TeamName      string  `json:"team_name"`
	Points        int     `json:"points"`
	SpeakerPoints float64 `json:"speaker_points"`
	Margin        float64 `json:"margin"`
	IsNovice      bool    `json:"is_novice"`
	IsEsl         bool    `json:"is_esl"`
	IsEfl         bool    `json:"is_efl"`
	Bubble        bool    `json:"bubble"`
}

// BreakResult is the computed (or persisted) qualifier list for one break category.
type BreakResult struct {
	CategoryID   string      `json:"category_id"`
	CategoryName string      `json:"category_name"`
	Size         *int        `json:"size"`
	Cutoff       int         `json:"cutoff"`
	Qualifiers   []BreakTeam `json:"qualifiers"`
}

// BracketDebate is one knockout debate; Bye debates carry a single team.
type BracketDebate struct {
	ID              string           `json:"id"`
	Venue           string           `json:"venue"`
	BracketPosition *int             `json:"bracket_position"`
	Bye             bool             `json:"bye"`
	Teams           []TeamAssignment `json:"teams"`
	WinnerTeamID    *string          `json:"winner_team_id"`
}

// BracketRound is one elimination round in the bracket visualizer.
type BracketRound struct {
	ID      string          `json:"id"`
	Seq     int             `json:"seq"`
	Name    string          `json:"name"`
	Debates []BracketDebate `json:"debates"`
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

// Conflict represents a declared clash between two draw entities.
type Conflict struct {
	ID          string `json:"id"`
	SubjectType string `json:"subject_type"` // 'adjudicator', 'team'
	SubjectID   string `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	TargetType  string `json:"target_type"` // 'team', 'speaker', 'adjudicator', 'institution'
	TargetID    string `json:"target_id"`
	TargetName  string `json:"target_name"`
	Weight      string `json:"weight"` // 'hard', 'soft'
}

// FeedbackQuestion represents a configurable feedback questionnaire item.
type FeedbackQuestion struct {
	ID       string   `json:"id"`
	Seq      int      `json:"seq"`
	Type     string   `json:"type"` // 'scale', 'text', 'checkbox', 'select'
	Name     string   `json:"name"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
	FromType string   `json:"from_type"` // 'team', 'adjudicator'
	ToType   string   `json:"to_type"`   // 'adjudicator'
}

// FeedbackSubmission represents a submitted feedback answer set for an adjudicator.
type FeedbackSubmission struct {
	ID         string            `json:"id"`
	RoundID    string            `json:"round_id"`
	DebateID   string            `json:"debate_id"`
	SourceType string            `json:"source_type"` // 'team', 'adjudicator'
	SourceID   string            `json:"source_id"`
	SourceName string            `json:"source_name"`
	TargetID   string            `json:"target_adjudicator_id"`
	TargetName string            `json:"target_name"`
	Score      *float64          `json:"score"`
	CreatedAt  string            `json:"created_at"`
	Answers    map[string]string `json:"answers"`
}

// FeedbackTarget describes an adjudicator a token holder may evaluate in a given debate.
type FeedbackTarget struct {
	DebateID        string `json:"debate_id"`
	Venue           string `json:"venue"`
	RoundName       string `json:"round_name"`
	AdjudicatorID   string `json:"adjudicator_id"`
	AdjudicatorName string `json:"adjudicator_name"`
	Role            string `json:"role"` // 'chair', 'panel'
}

// Checkin represents the check-in state of a draw entity with its QR token.
type Checkin struct {
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	EntityName   string `json:"entity_name"`
	CheckedIn    bool   `json:"checked_in"`
	CheckinToken string `json:"checkin_token"`
}

// CheckinTokenInfo is the payload returned by check-in token endpoints;
// EntityID is internal bookkeeping and never serialized.
type CheckinTokenInfo struct {
	EntityType string `json:"entity_type"`
	EntityName string `json:"entity_name"`
	CheckedIn  bool   `json:"checked_in"`
	EntityID   string `json:"-"`
}

// AvailabilityOverride represents a single round availability row.
type AvailabilityOverride struct {
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	IsAvailable bool   `json:"is_available"`
}

// AvailabilityEntry describes one team/adjudicator's availability for a round;
// a nil IsAvailable means no override exists (default available).
type AvailabilityEntry struct {
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	Name        string `json:"name"`
	IsAvailable *bool  `json:"is_available"`
	CheckedIn   bool   `json:"checked_in"`
}

// Draw-specific models to communicate between draw algorithm and store:

type TeamDrawInfo struct {
	ID            string
	InstitutionID string
	Standby       bool
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
	DebateID        string
	Venue           string
	BracketPosition int
	Teams           []TeamAssignment
	Adjudicators    []AdjudicatorAssignment
}
