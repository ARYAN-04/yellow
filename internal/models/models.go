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
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	Seq                    int     `json:"seq"`
	Size                   *int    `json:"size"`
	BasePoints             *int    `json:"base_points"`
	MaxTeamsPerInstitution *int    `json:"max_teams_per_institution,omitempty"`
	Rule                   *string `json:"rule,omitempty"`
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
	ID       string    `json:"id,omitempty"`
	TeamID   string    `json:"team_id"`
	TeamName string    `json:"team_name"`
	Side     string    `json:"side"`
	PullUp   bool      `json:"pull_up,omitempty"`
	Speakers []Speaker `json:"speakers,omitempty"`
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
	ID              string                  `json:"id"`
	Venue           string                  `json:"venue"`
	VenueAccessible bool                    `json:"venue_accessible,omitempty"`
	Teams           []TeamAssignment        `json:"teams"`
	Adjudicators    []AdjudicatorAssignment `json:"adjudicators"`
}

// SpeakerScore represents a recorded individual speaker score on a ballot.
type SpeakerScore struct {
	ID          string  `json:"id"`
	BallotID    string  `json:"ballot_id"`
	SpeakerID   string  `json:"speaker_id"`
	SpeakerName string  `json:"speaker_name,omitempty"`
	TeamID      string  `json:"team_id"`
	Score       float64 `json:"score"`
	IsReply     bool    `json:"is_reply"`
	SpeechOrder int     `json:"speech_order"`
	Role        string  `json:"role,omitempty"`
}

// SpeakerScoreInput represents an individual speaker score input in a ballot payload.
type SpeakerScoreInput struct {
	SpeakerID   string  `json:"speaker_id"`
	Score       float64 `json:"score"`
	IsReply     bool    `json:"is_reply"`
	SpeechOrder int     `json:"speech_order"`
	Role        string  `json:"role,omitempty"`
}

// TeamBallotResult represents the result points for a team.
// AdjudicatorID is set on split (per-judge) ballots; NULL means consensus.
type TeamBallotResult struct {
	TeamID        string              `json:"team_id"`
	Points        int                 `json:"points"`
	SpeakerPoints float64             `json:"speaker_points"`
	AdjudicatorID *string             `json:"adjudicator_id,omitempty"`
	SpeakerScores []SpeakerScoreInput `json:"speaker_scores,omitempty"`
}

// SpeakerStanding represents an individual speaker's cumulative rank and statistics.
type SpeakerStanding struct {
	Rank            int     `json:"rank"`
	SpeakerID       string  `json:"speaker_id"`
	SpeakerName     string  `json:"speaker_name"`
	TeamID          string  `json:"team_id"`
	TeamName        string  `json:"team_name"`
	InstitutionCode string  `json:"institution_code"`
	TotalScore      float64 `json:"total_score"`
	AverageScore    float64 `json:"average_score"`
	TrimmedScore    float64 `json:"trimmed_score"`
	SpeechCount     int     `json:"speech_count"`
	IsNovice        bool    `json:"is_novice"`
	IsEsl           bool    `json:"is_esl"`
	IsEfl           bool    `json:"is_efl"`
}

// AdjudicatorStanding represents an adjudicator's cumulative rankings, debate counts, and feedback metrics.
type AdjudicatorStanding struct {
	Rank                 int      `json:"rank"`
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	InstitutionID        *string  `json:"institution_id,omitempty"`
	InstitutionName      *string  `json:"institution_name,omitempty"`
	InstitutionCode      *string  `json:"institution_code,omitempty"`
	TestScore            float64  `json:"test_score"`
	FeedbackRating       *float64 `json:"feedback_rating,omitempty"`
	DebatesCount         int      `json:"debates_count"`
	ChairsCount          int      `json:"chairs_count"`
	PanelsCount          int      `json:"panels_count"`
	TraineesCount        int      `json:"trainees_count"`
	AverageFeedbackScore *float64 `json:"average_feedback_score,omitempty"`
	FeedbackCount        int      `json:"feedback_count"`
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
	Slug      string    `json:"slug"`
	Type      string    `json:"type"` // 'team', 'adjudicator'
	OwnerID   string    `json:"owner_id"`
	OwnerName string    `json:"owner_name"`
	Speakers  []Speaker `json:"speakers,omitempty"`
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
	Motion        string           `json:"motion,omitempty"`
	Chair         string           `json:"chair,omitempty"`
	Panellists    []string         `json:"panellists,omitempty"`
	Adjudicators  []string         `json:"adjudicators,omitempty"`
	SpeakerScores []SpeakerScore   `json:"speaker_scores,omitempty"`
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

// Motion represents a motion and its info slide for a tournament round.
type Motion struct {
	ID         string  `json:"id"`
	RoundID    string  `json:"round_id"`
	Seq        int     `json:"seq"`
	Reference  string  `json:"reference"`
	Text       string  `json:"text"`
	InfoSlide  string  `json:"info_slide"`
	ReleasedAt *string `json:"released_at"`
}

// MotionVeto represents a team's preference or veto ranking for a motion in a debate.
type MotionVeto struct {
	ID              string `json:"id"`
	DebateID        string `json:"debate_id"`
	TeamID          string `json:"team_id"`
	TeamName        string `json:"team_name"`
	MotionID        string `json:"motion_id"`
	MotionReference string `json:"motion_reference"`
	MotionText      string `json:"motion_text"`
	Preference      int    `json:"preference"`
}

// MotionStatistics represents win rate and positional statistics for a motion.
type MotionStatistics struct {
	MotionID         string                 `json:"motion_id"`
	Reference        string                 `json:"reference"`
	Text             string                 `json:"text"`
	RoundName        string                 `json:"round_name"`
	TotalDebates     int                    `json:"total_debates"`
	SideWins         map[string]int         `json:"side_wins"`
	SidePercentages  map[string]float64     `json:"side_percentages"`
	PositionalCounts map[string]map[int]int `json:"positional_counts"`
}

// Venue represents a debate venue/room with priority and accessibility flags.
type Venue struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Priority     int     `json:"priority"`
	CategoryID   *string `json:"category_id"`
	IsAccessible bool    `json:"is_accessible"`
}

// TeamTrajectoryOpponent represents an opposing team in a debate.
type TeamTrajectoryOpponent struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	Side     string `json:"side"`
}

// TeamTrajectoryAdjudicator represents an adjudicator assigned to a debate.
type TeamTrajectoryAdjudicator struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// SpeakerScoreEntry represents a single speaker score entry on a ballot.
type SpeakerScoreEntry struct {
	SpeakerID   string  `json:"speaker_id"`
	SpeakerName string  `json:"speaker_name"`
	Score       float64 `json:"score"`
	IsReply     bool    `json:"is_reply"`
	SpeechOrder int     `json:"speech_order"`
	Role        string  `json:"role,omitempty"`
}

// TeamTrajectoryDebate represents a single round's debate in a team's trajectory.
type TeamTrajectoryDebate struct {
	RoundID         string                      `json:"round_id"`
	RoundSeq        int                         `json:"round_seq"`
	RoundName       string                      `json:"round_name"`
	RoundStage      string                      `json:"round_stage"`
	Silent          bool                        `json:"silent"`
	DrawReleased    bool                        `json:"draw_released"`
	ResultsReleased bool                        `json:"results_released"`
	DebateID        string                      `json:"debate_id"`
	Venue           string                      `json:"venue"`
	Side            string                      `json:"side"`
	PullUp          bool                        `json:"pull_up"`
	Opponents       []TeamTrajectoryOpponent    `json:"opponents"`
	Adjudicators    []TeamTrajectoryAdjudicator `json:"adjudicators"`
	BallotStatus    *string                     `json:"ballot_status,omitempty"`
	Points          *int                        `json:"points,omitempty"`
	SpeakerPoints   *float64                    `json:"speaker_points,omitempty"`
	SpeakerScores   []SpeakerScoreEntry         `json:"speaker_scores,omitempty"`
}

// TeamTrajectory represents a team's full trajectory and round-by-round history.
type TeamTrajectory struct {
	Team    Team                   `json:"team"`
	Debates []TeamTrajectoryDebate `json:"debates"`
}

// SpeakerTrajectorySpeech represents a speech in a speaker's trajectory.
type SpeakerTrajectorySpeech struct {
	RoundID         string   `json:"round_id"`
	RoundSeq        int      `json:"round_seq"`
	RoundName       string   `json:"round_name"`
	RoundStage      string   `json:"round_stage"`
	Silent          bool     `json:"silent"`
	DrawReleased    bool     `json:"draw_released"`
	ResultsReleased bool     `json:"results_released"`
	DebateID        string   `json:"debate_id"`
	Venue           string   `json:"venue"`
	Side            string   `json:"side"`
	SpeechOrder     int      `json:"speech_order"`
	IsReply         bool     `json:"is_reply"`
	Role            string   `json:"role,omitempty"`
	Score           *float64 `json:"score,omitempty"`
	TeamPoints      *int     `json:"team_points,omitempty"`
	BallotStatus    *string  `json:"ballot_status,omitempty"`
}

// SpeakerTrajectory represents a speaker's full trajectory and speech history.
type SpeakerTrajectory struct {
	Speaker  Speaker                   `json:"speaker"`
	TeamID   string                    `json:"team_id"`
	TeamName string                    `json:"team_name"`
	Speeches []SpeakerTrajectorySpeech `json:"speeches"`
}

// AdjudicatorTrajectoryDebate represents a single round's debate in an adjudicator's judging history.
type AdjudicatorTrajectoryDebate struct {
	RoundID         string                      `json:"round_id"`
	RoundSeq        int                         `json:"round_seq"`
	RoundName       string                      `json:"round_name"`
	RoundStage      string                      `json:"round_stage"`
	Silent          bool                        `json:"silent"`
	DrawReleased    bool                        `json:"draw_released"`
	ResultsReleased bool                        `json:"results_released"`
	DebateID        string                      `json:"debate_id"`
	Venue           string                      `json:"venue"`
	Role            string                      `json:"role"`
	Teams           []TeamTrajectoryOpponent    `json:"teams"`
	CoAdjudicators  []TeamTrajectoryAdjudicator `json:"co_adjudicators"`
	BallotStatus    *string                     `json:"ballot_status,omitempty"`
}

// AdjudicatorTrajectory represents an adjudicator's full judging history.
type AdjudicatorTrajectory struct {
	Adjudicator Adjudicator                   `json:"adjudicator"`
	Debates     []AdjudicatorTrajectoryDebate `json:"debates"`
}
