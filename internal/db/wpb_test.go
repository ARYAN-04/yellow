package db

import (
	"testing"

	"yellow/internal/models"

	"github.com/google/uuid"
)

func TestMotionsCRUD(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	r1ID := uuid.New().String()
	if err := store.CreateRound(r1ID, 1, "Round 1", "preliminary"); err != nil {
		t.Fatal(err)
	}

	m1 := models.Motion{
		ID:        uuid.New().String(),
		RoundID:   r1ID,
		Seq:       1,
		Reference: "R1-M1",
		Text:      "This House would ban private schools.",
		InfoSlide: "Private schools are independent institutions funded by tuition fees.",
	}
	m2 := models.Motion{
		ID:        uuid.New().String(),
		RoundID:   r1ID,
		Seq:       2,
		Reference: "R1-M2",
		Text:      "This House regrets the rise of social media influencers.",
	}

	if err := store.CreateMotion(m1); err != nil {
		t.Fatalf("failed to create motion 1: %v", err)
	}
	if err := store.CreateMotion(m2); err != nil {
		t.Fatalf("failed to create motion 2: %v", err)
	}

	// List motions
	list, err := store.ListMotions(r1ID)
	if err != nil {
		t.Fatalf("failed to list motions: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 motions, got %d", len(list))
	}
	if list[0].Reference != "R1-M1" || list[0].InfoSlide != m1.InfoSlide {
		t.Errorf("motion 1 metadata mismatch: %+v", list[0])
	}
	if list[0].ReleasedAt != nil {
		t.Errorf("expected motion 1 unreleased initially, got %v", *list[0].ReleasedAt)
	}

	// Release motions
	if err := store.ReleaseMotions(r1ID, true); err != nil {
		t.Fatalf("failed to release motions: %v", err)
	}
	list, err = store.ListMotions(r1ID)
	if err != nil || len(list) != 2 || list[0].ReleasedAt == nil {
		t.Fatalf("expected motions released, got %+v", list)
	}

	// Unrelease
	if err := store.ReleaseMotions(r1ID, false); err != nil {
		t.Fatalf("failed to unrelease motions: %v", err)
	}
	list, err = store.ListMotions(r1ID)
	if err != nil || len(list) != 2 || list[0].ReleasedAt != nil {
		t.Fatalf("expected motions unreleased, got %+v", list)
	}

	// Update motion
	m1.Text = "This House would abolish all private schools completely."
	if err := store.UpdateMotion(m1); err != nil {
		t.Fatalf("failed to update motion: %v", err)
	}
	list, _ = store.ListMotions(r1ID)
	if list[0].Text != m1.Text {
		t.Errorf("expected updated text %q, got %q", m1.Text, list[0].Text)
	}

	// Delete motion
	if err := store.DeleteMotion(m2.ID); err != nil {
		t.Fatalf("failed to delete motion: %v", err)
	}
	list, _ = store.ListMotions(r1ID)
	if len(list) != 1 {
		t.Fatalf("expected 1 motion remaining, got %d", len(list))
	}
}

func TestMotionVetoesAndPreferences(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	rID := uuid.New().String()
	_ = store.CreateRound(rID, 1, "Round 1", "preliminary")

	t1ID := uuid.New().String()
	t2ID := uuid.New().String()
	_ = store.CreateTeam(t1ID, "Team Alpha", "TA", "", []models.SpeakerRequest{{Name: "A1"}, {Name: "A2"}}, "tokA")
	_ = store.CreateTeam(t2ID, "Team Beta", "TB", "", []models.SpeakerRequest{{Name: "B1"}, {Name: "B2"}}, "tokB")

	m1ID := uuid.New().String()
	m2ID := uuid.New().String()
	m3ID := uuid.New().String()
	_ = store.CreateMotion(models.Motion{ID: m1ID, RoundID: rID, Seq: 1, Reference: "M1", Text: "Motion 1"})
	_ = store.CreateMotion(models.Motion{ID: m2ID, RoundID: rID, Seq: 2, Reference: "M2", Text: "Motion 2"})
	_ = store.CreateMotion(models.Motion{ID: m3ID, RoundID: rID, Seq: 3, Reference: "M3", Text: "Motion 3"})

	// Create debate
	debID := uuid.New().String()
	if err := store.SaveDraw(rID, []models.DebateDrawInput{
		{
			DebateID: debID,
			Venue:    "Room 101",
			Teams: []models.TeamAssignment{
				{TeamID: t1ID, Side: "Gov"},
				{TeamID: t2ID, Side: "Opp"},
			},
		},
	}); err != nil {
		t.Fatalf("failed to save draw: %v", err)
	}

	// Record vetoes: Team 1 prefers M1 (1), M2 (2), M3 (3)
	if err := store.RecordMotionVeto(debID, t1ID, m1ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordMotionVeto(debID, t1ID, m2ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordMotionVeto(debID, t1ID, m3ID, 3); err != nil {
		t.Fatal(err)
	}

	// Team 2 prefers M2 (1), M1 (2), M3 (3)
	if err := store.RecordMotionVeto(debID, t2ID, m2ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordMotionVeto(debID, t2ID, m1ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordMotionVeto(debID, t2ID, m3ID, 3); err != nil {
		t.Fatal(err)
	}

	vetoes, err := store.GetDebateVetoes(debID)
	if err != nil {
		t.Fatalf("failed to get debate vetoes: %v", err)
	}
	if len(vetoes) != 6 {
		t.Fatalf("expected 6 veto records, got %d", len(vetoes))
	}

	// Verify upsert
	if err := store.RecordMotionVeto(debID, t1ID, m1ID, 2); err != nil {
		t.Fatal(err)
	}
	vetoes, _ = store.GetDebateVetoes(debID)
	if len(vetoes) != 6 {
		t.Fatalf("expected still 6 veto records after upsert, got %d", len(vetoes))
	}
}

func TestMotionStatisticsCalculation(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	r1ID := uuid.New().String()
	r2ID := uuid.New().String()
	_ = store.CreateRound(r1ID, 1, "Round 1", "preliminary")
	_ = store.CreateRound(r2ID, 2, "Round 2", "preliminary")

	// Teams
	t1ID := uuid.New().String()
	t2ID := uuid.New().String()
	t3ID := uuid.New().String()
	t4ID := uuid.New().String()
	_ = store.CreateTeam(t1ID, "Team 1", "T1", "", []models.SpeakerRequest{{Name: "S1"}}, "tok1")
	_ = store.CreateTeam(t2ID, "Team 2", "T2", "", []models.SpeakerRequest{{Name: "S2"}}, "tok2")
	_ = store.CreateTeam(t3ID, "Team 3", "T3", "", []models.SpeakerRequest{{Name: "S3"}}, "tok3")
	_ = store.CreateTeam(t4ID, "Team 4", "T4", "", []models.SpeakerRequest{{Name: "S4"}}, "tok4")

	// Motions
	m1ID := uuid.New().String()
	m2ID := uuid.New().String()
	_ = store.CreateMotion(models.Motion{ID: m1ID, RoundID: r1ID, Seq: 1, Reference: "R1-M1", Text: "Motion for Round 1"})
	_ = store.CreateMotion(models.Motion{ID: m2ID, RoundID: r2ID, Seq: 1, Reference: "R2-M1", Text: "Motion for Round 2"})

	// Debates for Round 1 (BP format)
	deb1ID := uuid.New().String()
	_ = store.SaveDraw(r1ID, []models.DebateDrawInput{
		{
			DebateID: deb1ID,
			Venue:    "Room 1",
			Teams: []models.TeamAssignment{
				{TeamID: t1ID, Side: "OG"},
				{TeamID: t2ID, Side: "OO"},
				{TeamID: t3ID, Side: "CG"},
				{TeamID: t4ID, Side: "CO"},
			},
		},
	})

	// Submit and confirm ballot for Round 1: OG gets 3 pts (1st), OO gets 2 pts (2nd), CG gets 1 pt (3rd), CO gets 0 pts (4th)
	b1ID := uuid.New().String()
	if err := store.SubmitBallot(deb1ID, b1ID, "organizer", "admin", "submitted", false, "", []models.TeamBallotResult{
		{TeamID: t1ID, Points: 3, SpeakerPoints: 155},
		{TeamID: t2ID, Points: 2, SpeakerPoints: 150},
		{TeamID: t3ID, Points: 1, SpeakerPoints: 145},
		{TeamID: t4ID, Points: 0, SpeakerPoints: 140},
	}); err != nil {
		t.Fatalf("failed to submit ballot: %v", err)
	}
	if err := store.ConfirmBallot(b1ID); err != nil {
		t.Fatalf("failed to confirm ballot: %v", err)
	}

	// Compute statistics
	stats, err := store.GetMotionStatistics()
	if err != nil {
		t.Fatalf("failed to get motion statistics: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 motion statistics entries, got %d", len(stats))
	}

	// Find stat for M1
	var statM1 *models.MotionStatistics
	for i := range stats {
		if stats[i].MotionID == m1ID {
			statM1 = &stats[i]
		}
	}
	if statM1 == nil {
		t.Fatal("expected stat for M1")
	}

	if statM1.TotalDebates != 1 {
		t.Errorf("expected 1 debate for M1, got %d", statM1.TotalDebates)
	}
	if statM1.SideWins["OG"] != 1 {
		t.Errorf("expected OG to have 1 win, got %d", statM1.SideWins["OG"])
	}
	if statM1.SidePercentages["OG"] != 100.0 {
		t.Errorf("expected OG percentage 100.0, got %v", statM1.SidePercentages["OG"])
	}
	if statM1.PositionalCounts["OG"][1] != 1 {
		t.Errorf("expected OG 1st place count 1, got %d", statM1.PositionalCounts["OG"][1])
	}
	if statM1.PositionalCounts["OO"][2] != 1 {
		t.Errorf("expected OO 2nd place count 1, got %d", statM1.PositionalCounts["OO"][2])
	}
}

func TestVenuesCRUDAndAvailability(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	v1 := models.Venue{
		ID:           uuid.New().String(),
		Name:         "Main Auditorium",
		Priority:     100,
		IsAccessible: true,
	}
	v2 := models.Venue{
		ID:           uuid.New().String(),
		Name:         "Room B-101",
		Priority:     50,
		IsAccessible: false,
	}
	v3 := models.Venue{
		ID:           uuid.New().String(),
		Name:         "Room C-202",
		Priority:     10,
		IsAccessible: true,
	}

	if err := store.CreateVenue(v1); err != nil {
		t.Fatalf("failed to create venue 1: %v", err)
	}
	if err := store.CreateVenue(v2); err != nil {
		t.Fatalf("failed to create venue 2: %v", err)
	}
	if err := store.CreateVenue(v3); err != nil {
		t.Fatalf("failed to create venue 3: %v", err)
	}

	// List venues (should be sorted by priority DESC)
	list, err := store.ListVenues()
	if err != nil {
		t.Fatalf("failed to list venues: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 venues, got %d", len(list))
	}
	if list[0].Name != "Main Auditorium" || list[1].Name != "Room B-101" || list[2].Name != "Room C-202" {
		t.Errorf("unexpected venue order: %+v", list)
	}

	// Update venue
	v2.Priority = 120
	if err := store.UpdateVenue(v2); err != nil {
		t.Fatalf("failed to update venue: %v", err)
	}
	list, _ = store.ListVenues()
	if list[0].Name != "Room B-101" {
		t.Errorf("expected Room B-101 to now have highest priority, got %s", list[0].Name)
	}

	// Availability for round
	rID := uuid.New().String()
	_ = store.CreateRound(rID, 1, "Round 1", "preliminary")

	// Mark Room B-101 unavailable for round 1
	if err := store.SetRoundAvailability(rID, "venue", v2.ID, false); err != nil {
		t.Fatalf("failed to set round availability: %v", err)
	}

	avail, err := store.GetAvailableVenuesForRound(rID)
	if err != nil {
		t.Fatalf("failed to get available venues: %v", err)
	}
	if len(avail) != 2 {
		t.Fatalf("expected 2 available venues, got %d", len(avail))
	}
	for _, v := range avail {
		if v.ID == v2.ID {
			t.Errorf("expected venue %s to be excluded from round availability", v2.Name)
		}
	}

	// CSV Import
	imported, err := store.ImportVenues([]models.Venue{
		{Name: "Imported Hall 1", Priority: 80, IsAccessible: true},
		{Name: "Imported Hall 2", Priority: 40, IsAccessible: false},
	})
	if err != nil {
		t.Fatalf("failed to import venues: %v", err)
	}
	if imported != 2 {
		t.Errorf("expected 2 imported venues, got %d", imported)
	}

	// Delete venue
	if err := store.DeleteVenue(v3.ID); err != nil {
		t.Fatalf("failed to delete venue: %v", err)
	}
}
