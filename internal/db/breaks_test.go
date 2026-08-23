package db

import (
	"testing"

	"yellow/internal/models"

	"github.com/google/uuid"
)

func TestNextPow2(t *testing.T) {
	cases := map[int]int{1: 1, 2: 2, 3: 4, 8: 8, 9: 16, 12: 16}
	for in, want := range cases {
		if got := nextPow2(in); got != want {
			t.Errorf("nextPow2(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestEliminationName(t *testing.T) {
	cases := map[int]string{2: "Final", 4: "Semifinals", 8: "Quarterfinals", 16: "Octofinals", 6: "Elimination 6"}
	for in, want := range cases {
		if got := eliminationName(in); got != want {
			t.Errorf("eliminationName(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildKnockoutDebatesEightSeeds(t *testing.T) {
	seeds := []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}
	sides := []string{"OG", "OO"}
	debates := buildKnockoutDebates(seeds, sides)

	if len(debates) != 4 {
		t.Fatalf("got %d debates, want 4", len(debates))
	}
	wantPairs := [][2]string{{"s1", "s8"}, {"s2", "s7"}, {"s3", "s6"}, {"s4", "s5"}}
	for i, d := range debates {
		if d.BracketPosition != i+1 {
			t.Errorf("debate %d position = %d, want %d", i, d.BracketPosition, i+1)
		}
		if len(d.Teams) != 2 {
			t.Fatalf("debate %d has %d teams, want 2", i, len(d.Teams))
		}
		if d.Teams[0].TeamID != wantPairs[i][0] || d.Teams[0].Side != "OG" {
			t.Errorf("debate %d team A = %s/%s, want %s/OG", i, d.Teams[0].TeamID, d.Teams[0].Side, wantPairs[i][0])
		}
		if d.Teams[1].TeamID != wantPairs[i][1] || d.Teams[1].Side != "OO" {
			t.Errorf("debate %d team B = %s/%s, want %s/OO", i, d.Teams[1].TeamID, d.Teams[1].Side, wantPairs[i][1])
		}
	}
}

func TestBuildKnockoutDebatesByes(t *testing.T) {
	seeds := []string{"a", "b", "c", "d", "e"}
	sides := []string{"Gov", "Opp"}
	debates := buildKnockoutDebates(seeds, sides)

	if len(debates) != 4 {
		t.Fatalf("got %d debates, want 4", len(debates))
	}
	type pair struct {
		a, b string
		bye  bool
	}
	want := []pair{{"a", "", true}, {"b", "", true}, {"c", "", true}, {"d", "e", false}}
	for i, w := range want {
		d := debates[i]
		got := pair{bye: len(d.Teams) == 1}
		for _, tm := range d.Teams {
			if tm.Side == sides[0] {
				got.a = tm.TeamID
			} else if tm.Side == sides[1] {
				got.b = tm.TeamID
			}
		}
		if got.a != w.a || got.b != w.b || got.bye != w.bye {
			t.Errorf("debate %d = %+v (%d teams), want %s/%s bye=%v", i, got, len(d.Teams), w.a, w.b, w.bye)
		}
	}
}

func TestPickWinner(t *testing.T) {
	teamIDs := []string{"t1", "t2"}
	scores := map[string]map[string]debateScore{
		"d1": {"t1": {points: 2}, "t2": {points: 1}},
		"d2": {"t1": {points: 1, speakerPoints: 250}, "t2": {points: 1, speakerPoints: 255}},
		"d3": {"t1": {points: 1, speakerPoints: 250}, "t2": {points: 1, speakerPoints: 250}},
		"d4": {"t1": {points: 2}},
	}
	if w, ok := pickWinner(scores["d1"], teamIDs); !ok || w != "t1" {
		t.Errorf("d1 winner = %q ok=%v, want t1", w, ok)
	}
	if w, ok := pickWinner(scores["d2"], teamIDs); !ok || w != "t2" {
		t.Errorf("d2 winner = %q ok=%v, want t2 on speaker points", w, ok)
	}
	if _, ok := pickWinner(scores["d3"], teamIDs); ok {
		t.Error("d3 tie should be undecided")
	}
	if _, ok := pickWinner(scores["d4"], teamIDs); ok {
		t.Error("d4 missing result should be undecided")
	}
}

func TestComputeBreakAndBracketRoundTrip(t *testing.T) {
	store, err := NewSQLiteStore(t.TempDir() + "/wp8.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	teams := []struct {
		name   string
		novice bool
	}{
		{"Alpha", false}, {"Beta", true}, {"Gamma", false}, {"Delta", true},
		{"Epsilon", false}, {"Zeta", false}, {"Eta", true}, {"Theta", false},
	}
	for _, tm := range teams {
		id := uuid.New().String()
		if err := store.CreateTeam(id, tm.name, "", "", nil, uuid.New().String()); err != nil {
			t.Fatal(err)
		}
		if tm.novice {
			yes := true
			if err := store.UpdateTeam(id, nil, nil, nil, &yes, nil, nil, nil); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := store.CreateBreakCategory(models.BreakCategory{ID: "cat1", Name: "Novice Cup", Seq: 1}); err != nil {
		t.Fatal(err)
	}
	if err := store.SetConfig("ranking_precedence", `["points","speaker_points"]`); err != nil {
		t.Fatal(err)
	}

	res, err := store.ComputeBreak("open")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Qualifiers) != 8 || res.Cutoff != 0 {
		t.Fatalf("open break = %d qualifiers cutoff %d, want 8/0", len(res.Qualifiers), res.Cutoff)
	}

	res, err = store.ComputeBreak("novice")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Qualifiers) != 3 {
		t.Fatalf("novice break = %d qualifiers, want 3", len(res.Qualifiers))
	}
	for _, q := range res.Qualifiers {
		if !q.IsNovice {
			t.Errorf("non-novice team %s in novice break", q.TeamName)
		}
	}

	size := 2
	if err := store.UpdateBreakCategory(models.BreakCategory{ID: "cat1", Name: "Novice Cup", Seq: 1, Size: &size}); err != nil {
		t.Fatal(err)
	}
	res, err = store.ComputeBreak("cat1")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Qualifiers) < 2 || res.Qualifiers[0].Rank != 1 {
		t.Fatalf("unexpected cat1 break: %+v", res.Qualifiers)
	}
	if res.Qualifiers[len(res.Qualifiers)-1].Rank <= size {
		t.Errorf("bubble teams missing beyond size cap")
	}
	last := res.Qualifiers[len(res.Qualifiers)-1]
	if last.Points != res.Cutoff || !last.Bubble {
		t.Errorf("last row points=%d bubble=%v, want tied at cutoff %d with bubble flag", last.Points, last.Bubble, res.Cutoff)
	}

	if err := store.SaveBreakSnapshot("open", res.Qualifiers[:2]); err != nil {
		t.Fatal(err)
	}
	seeds, err := store.getBreakSeeds("open")
	if err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 2 {
		t.Fatalf("snapshot seeds = %v, want 2", seeds)
	}
}
