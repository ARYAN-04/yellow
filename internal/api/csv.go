package api

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	"GoTabs/internal/models"
)

// ImportInstitutions handles CSV uploads to insert institutions in bulk.
func (api *API) ImportInstitutions(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		JSONError(w, "file field is required in form-data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		JSONError(w, "failed to read CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		JSONError(w, "CSV is empty or missing headers", http.StatusBadRequest)
		return
	}

	// Read headers and map index
	header := records[0]
	nameIdx, codeIdx := -1, -1
	for idx, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "name" {
			nameIdx = idx
		} else if h == "code" {
			codeIdx = idx
		}
	}

	if nameIdx == -1 || codeIdx == -1 {
		JSONError(w, "missing required headers: 'name' and 'code' must be present", http.StatusBadRequest)
		return
	}

	var insts []models.Institution
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= nameIdx || len(row) <= codeIdx {
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		code := strings.ToUpper(strings.TrimSpace(row[codeIdx]))
		if name == "" || code == "" {
			continue
		}

		insts = append(insts, models.Institution{
			Name: name,
			Code: code,
		})
	}

	inserted, err := tdb.ImportInstitutions(insts)
	if err != nil {
		JSONError(w, "failed to import institutions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]interface{}{"status": "success", "imported": inserted}, http.StatusOK)
}

// ImportTeams handles CSV uploads to insert teams and speakers in bulk.
func (api *API) ImportTeams(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		JSONError(w, "file field is required in form-data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		JSONError(w, "failed to read CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		JSONError(w, "CSV is empty or missing headers", http.StatusBadRequest)
		return
	}

	header := records[0]
	nameIdx, codeIdx, instIdx := -1, -1, -1
	var speakerIndices []int

	for idx, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "name" {
			nameIdx = idx
		} else if h == "code" {
			codeIdx = idx
		} else if h == "institution_code" || h == "institution" {
			instIdx = idx
		} else if strings.HasPrefix(h, "speaker") {
			speakerIndices = append(speakerIndices, idx)
		}
	}

	if nameIdx == -1 {
		JSONError(w, "missing required header: 'name' must be present", http.StatusBadRequest)
		return
	}

	var teams []models.TeamImport
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= nameIdx {
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			continue
		}

		code := ""
		if codeIdx != -1 && len(row) > codeIdx {
			code = strings.TrimSpace(row[codeIdx])
		}

		instCode := ""
		if instIdx != -1 && len(row) > instIdx && row[instIdx] != "" {
			instCode = strings.TrimSpace(row[instIdx])
		}

		var speakers []string
		for _, spIdx := range speakerIndices {
			if len(row) > spIdx && row[spIdx] != "" {
				speakers = append(speakers, strings.TrimSpace(row[spIdx]))
			}
		}

		teams = append(teams, models.TeamImport{
			Name:            name,
			Code:            code,
			InstitutionCode: instCode,
			Speakers:        speakers,
		})
	}

	inserted, err := tdb.ImportTeams(teams)
	if err != nil {
		JSONError(w, "failed to import teams: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]interface{}{"status": "success", "imported": inserted}, http.StatusOK)
}

// ImportAdjudicators handles CSV uploads to insert adjudicators in bulk.
func (api *API) ImportAdjudicators(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		JSONError(w, "file field is required in form-data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		JSONError(w, "failed to read CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		JSONError(w, "CSV is empty or missing headers", http.StatusBadRequest)
		return
	}

	header := records[0]
	nameIdx, instIdx, scoreIdx := -1, -1, -1

	for idx, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "name" {
			nameIdx = idx
		} else if h == "institution_code" || h == "institution" {
			instIdx = idx
		} else if h == "test_score" || h == "score" {
			scoreIdx = idx
		}
	}

	if nameIdx == -1 {
		JSONError(w, "missing required header: 'name' must be present", http.StatusBadRequest)
		return
	}

	var adjs []models.AdjudicatorImport
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= nameIdx {
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			continue
		}

		instCode := ""
		if instIdx != -1 && len(row) > instIdx && row[instIdx] != "" {
			instCode = strings.TrimSpace(row[instIdx])
		}

		score := 0.0
		if scoreIdx != -1 && len(row) > scoreIdx && row[scoreIdx] != "" {
			parsedScore, err := strconv.ParseFloat(strings.TrimSpace(row[scoreIdx]), 64)
			if err == nil {
				score = parsedScore
			}
		}

		adjs = append(adjs, models.AdjudicatorImport{
			Name:            name,
			InstitutionCode: instCode,
			TestScore:       score,
		})
	}

	inserted, err := tdb.ImportAdjudicators(adjs)
	if err != nil {
		JSONError(w, "failed to import adjudicators: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]interface{}{"status": "success", "imported": inserted}, http.StatusOK)
}
