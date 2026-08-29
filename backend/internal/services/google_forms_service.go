package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/google"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

type FieldMap struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

var allowedTargets = map[string]bool{
	"first_name": true, "last_name": true, "email": true, "phone": true,
	"address": true, "date_of_birth": true, "education": true, "experience": true,
	"skills": true, "resume_path": true, "source": true, "notes": true,
	"status": true, "cover_letter": true, "response_id": true, "submitted_at": true,
}

var defaultFieldRules = []FieldMap{
	{Source: "Full Name", Target: "first_name"},
	{Source: "Name", Target: "first_name"},
	{Source: "First Name", Target: "first_name"},
	{Source: "Last Name", Target: "last_name"},
	{Source: "Email", Target: "email"},
	{Source: "Email Address", Target: "email"},
	{Source: "Phone", Target: "phone"},
	{Source: "Phone Number", Target: "phone"},
	{Source: "Contact", Target: "phone"},
	{Source: "Address", Target: "address"},
	{Source: "Date of Birth", Target: "date_of_birth"},
	{Source: "DOB", Target: "date_of_birth"},
	{Source: "Birthdate", Target: "date_of_birth"},
	{Source: "Education", Target: "education"},
	{Source: "Qualification", Target: "education"},
	{Source: "Experience", Target: "experience"},
	{Source: "Years of Experience", Target: "experience"},
	{Source: "Skills", Target: "skills"},
	{Source: "Technical Skills", Target: "skills"},
	{Source: "Resume/CV", Target: "resume_path"},
	{Source: "Resume URL", Target: "resume_path"},
	{Source: "Resume Link", Target: "resume_path"},
	{Source: "CV", Target: "resume_path"},
	{Source: "CV Link", Target: "resume_path"},
	{Source: "Cover Letter", Target: "cover_letter"},
	{Source: "Response ID", Target: "response_id"},
	{Source: "ID", Target: "response_id"},
	{Source: "Timestamp", Target: "submitted_at"},
}

type GoogleFormInput struct {
	FormURL, SheetID, SheetName, Status, StatusDetail string
	HeaderRow                                         int
	FieldMapping                                      []FieldMap
}

type GoogleFormService struct {
	integrations *repositories.GoogleFormIntegrationRepository
	responses    *repositories.GoogleFormResponseRepository
	sheets       google.SheetsReader
	tokens       *google.TokenManager
	successURL   string
	recruit      *RecruitmentService
	notify       *notifications.Service
	audit        *auditmanager.Service
}

func NewGoogleFormService(
	integrations *repositories.GoogleFormIntegrationRepository,
	responses *repositories.GoogleFormResponseRepository,
	tokens *repositories.GoogleOAuthTokenRepository,
	sheets google.SheetsReader,
	ts *google.TokenManager,
	successURL string,
	recruit *RecruitmentService,
	notify *notifications.Service,
	audit *auditmanager.Service,
) *GoogleFormService {
	return &GoogleFormService{
		integrations: integrations,
		responses:    responses,
		sheets:       sheets,
		tokens:       ts,
		successURL:   successURL,
		recruit:      recruit,
		notify:       notify,
		audit:        audit,
	}
}

// --- Configuration ---

func (s *GoogleFormService) Connect(tenantID, jobID string, in GoogleFormInput, actorID, ip, ua string) (*models.GoogleFormIntegration, error) {
	if _, err := s.recruit.Job(tenantID, jobID); err != nil {
		return nil, ErrNotFound
	}
	if existing, _ := s.integrations.GetByJob(tenantID, jobID); existing != nil {
		return nil, ErrDuplicate
	}
	if in.FormURL == "" {
		return nil, fmt.Errorf("%w: google_form_url is required", ErrGoogleInvalidForm)
	}
	if !isHTTPURL(in.FormURL) {
		return nil, fmt.Errorf("%w: %q", ErrGoogleInvalidForm, in.FormURL)
	}
	if err := validateFieldMapping(in.FieldMapping); err != nil {
		return nil, err
	}
	status := orStr(in.Status, models.GoogleFormStatusConnected)
	if status == "" {
		status = models.GoogleFormStatusConnected
	}
	if in.SheetID == "" {
		status = models.GoogleFormStatusPending
	}
	integ := &models.GoogleFormIntegration{
		TenantID:      tenantID,
		JobID:         jobID,
		Provider:      "google_forms",
		FormURL:       in.FormURL,
		SpreadsheetID: in.SheetID,
		SheetName:     in.SheetName,
		HeaderRow:     in.HeaderRow,
		FieldMapping:  marshalFieldMapping(in.FieldMapping),
		Status:        status,
		StatusDetail:  in.StatusDetail,
	}
	if integ.HeaderRow < 1 {
		integ.HeaderRow = 1
	}
	if err := s.integrations.Create(integ); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "google_form_integration", integ.ID, ip, ua, map[string]string{"job_id": jobID})
	return integ, nil
}

func (s *GoogleFormService) Get(tenantID, jobID string) (*models.GoogleFormIntegration, error) {
	integ, err := s.integrations.GetByJob(tenantID, jobID)
	if err != nil {
		return nil, ErrNotFound
	}
	return integ, nil
}

func (s *GoogleFormService) Update(tenantID, jobID string, in GoogleFormInput, actorID, ip, ua string) (*models.GoogleFormIntegration, error) {
	integ, err := s.integrations.GetByJob(tenantID, jobID)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.FormURL != "" {
		if !isHTTPURL(in.FormURL) {
			return nil, fmt.Errorf("%w: %q", ErrGoogleInvalidForm, in.FormURL)
		}
	}
	if err := validateFieldMapping(in.FieldMapping); err != nil {
		return nil, err
	}
	fields := map[string]interface{}{}
	if in.FormURL != "" {
		fields["form_url"] = in.FormURL
	}
	if in.SheetID != "" {
		fields["spreadsheet_id"] = in.SheetID
	}
	if in.SheetName != "" {
		fields["sheet_name"] = in.SheetName
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if in.StatusDetail != "" {
		fields["status_detail"] = in.StatusDetail
	}
	if in.HeaderRow > 0 {
		fields["header_row"] = in.HeaderRow
	}
	if in.FieldMapping != nil {
		fields["field_mapping"] = marshalFieldMapping(in.FieldMapping)
	}
	if err := s.integrations.Update(tenantID, integ.ID, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "google_form_integration", integ.ID, ip, ua, nil)
	return s.integrations.GetByJob(tenantID, jobID)
}

func (s *GoogleFormService) Disconnect(tenantID, jobID, actorID, ip, ua string) error {
	integ, err := s.integrations.GetByJob(tenantID, jobID)
	if err != nil {
		return ErrNotFound
	}
	if err := s.integrations.DeleteByJob(tenantID, jobID); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "google_form_integration", integ.ID, ip, ua, map[string]string{"job_id": jobID})
	return nil
}

// --- OAuth ---

func (s *GoogleFormService) OAuthAuthorize() (string, string, error) {
	authURL, state, err := s.tokens.BeginAuth()
	if err != nil {
		return "", "", mapGoogleOAuthError(err)
	}
	return authURL, state, nil
}

func (s *GoogleFormService) OAuthCallback(code, state string) (string, error) {
	if err := s.tokens.Exchange(code, state); err != nil {
		return "", mapGoogleOAuthError(err)
	}
	if s.successURL == "" {
		return "/", nil
	}
	return s.successURL, nil
}

// --- Synchronization ---

func (s *GoogleFormService) Sync(tenantID, jobID, mode string, actorID, ip, ua string) (*dto.GoogleSyncResult, error) {
	integ, err := s.integrations.GetByJob(tenantID, jobID)
	if err != nil {
		return nil, ErrNotFound
	}
	if integ.Status == models.GoogleFormStatusDisconnected {
		return nil, ErrGoogleNotConnected
	}
	if integ.SpreadsheetID == "" {
		_ = s.integrations.Update(tenantID, integ.ID, map[string]interface{}{"sync_error": ErrGoogleInvalidSpreadsheet.Error(), "status": models.GoogleFormStatusError})
		return nil, fmt.Errorf("%w: google_sheet_id is not configured", ErrGoogleInvalidSpreadsheet)
	}

	sheet := integ.SheetName
	if sheet == "" {
		names, err := s.sheets.ListSheets(integ.SpreadsheetID)
		if err != nil {
			return nil, s.failSync(tenantID, integ, err)
		}
		if len(names) == 0 {
			return nil, s.failSync(tenantID, integ, ErrGoogleNoData)
		}
		sheet = names[0]
	}

	values, err := s.sheets.GetValues(integ.SpreadsheetID, sheet)
	if err != nil {
		return nil, s.failSync(tenantID, integ, err)
	}

	headerRow := integ.HeaderRow
	if headerRow < 1 {
		headerRow = 1
	}
	if len(values) < headerRow {
		return nil, s.failSync(tenantID, integ, ErrGoogleNoData)
	}

	rules, explicit, err := s.storedRules(integ.FieldMapping)
	if err != nil {
		return nil, s.failSync(tenantID, integ, err)
	}
	headerIdx, err := resolveFieldMapping(values[headerRow-1], rules, explicit)
	if err != nil {
		return nil, s.failSync(tenantID, integ, err)
	}
	if _, ok := headerIdx["email"]; !ok {
		return nil, s.failSync(tenantID, integ, ErrGoogleMissingEmail)
	}

	start := headerRow
	if mode != "full" && integ.SyncedRows > headerRow {
		start = integ.SyncedRows
	}
	if start < headerRow {
		start = headerRow
	}

	result := &dto.GoogleSyncResult{TotalRows: len(values) - headerRow}
	syncedRows := len(values)

	for i := start; i < len(values); i++ {
		row := values[i]
		if isBlankRow(row) {
			continue
		}
		fields := valuesFromRow(headerIdx, row)

		extID := externalResponseID(integ, i, fields["response_id"])
		exists, err := s.responses.ExistsByExternalID(tenantID, extID)
		if err != nil {
			_ = s.recordResponse(tenantID, integ, extID, nil, nil, nil, fields, models.GoogleFormResponseError, err.Error())
			result.Failed++
			continue
		}
		if exists {
			result.Duplicates++
			continue
		}

		email := fields["email"]
		if !validEmail(email) {
			_ = s.recordResponse(tenantID, integ, extID, nil, nil, nil, fields, models.GoogleFormResponseError, "missing or invalid email")
			result.Failed++
			continue
		}

		candidate, err := s.importCandidate(tenantID, actorID, ip, ua, fields)
		if err != nil {
			_ = s.recordResponse(tenantID, integ, extID, nil, nil, nil, fields, models.GoogleFormResponseError, err.Error())
			result.Failed++
			continue
		}

		submitted := parseSubmittedAt(fields["submitted_at"])
		app, err := s.recruit.CreateApplication(tenantID, ApplicationInput{
			JobPostID:   integ.JobID,
			CandidateID: candidate.ID,
			AppliedDate: appliedDate(submitted),
			CoverLetter: fields["cover_letter"],
		}, actorID, ip, ua)
		if err != nil {
			status := models.GoogleFormResponseError
			msg := err.Error()
			if errors.Is(err, ErrDuplicateApplication) {
				status = models.GoogleFormResponseDuplicate
				result.Duplicates++
			} else {
				result.Failed++
			}
			cID := candidate.ID
			_ = s.recordResponse(tenantID, integ, extID, &cID, nil, submitted, fields, status, msg)
			continue
		}

		result.Imported++
		cID := candidate.ID
		aID := app.ID
		_ = s.recordResponse(tenantID, integ, extID, &cID, &aID, submitted, fields, models.GoogleFormResponseImported, "")
	}

	_ = s.integrations.Update(tenantID, integ.ID, map[string]interface{}{
		"last_synced_at": time.Now(),
		"synced_rows":    syncedRows,
		"status":         models.GoogleFormStatusConnected,
		"sync_error":     "",
	})
	s.audit.Record(actorID, models.ActionUpdate, "google_form_sync", integ.ID, ip, ua, map[string]string{
		"job_id": jobID, "imported": strconv.Itoa(result.Imported),
	})
	if result.Failed > 0 || result.Duplicates > 0 {
		_ = s.notify.Notify(actorID, "Google Forms sync",
			fmt.Sprintf("Imported %d, skipped %d duplicate, failed %d for job %s.", result.Imported, result.Duplicates, result.Failed, integ.JobID),
			models.NotifyRecruitment, "", nil)
	}
	return result, nil
}

func (s *GoogleFormService) importCandidate(tenantID, actorID, ip, ua string, fields map[string]string) (*models.Candidate, error) {
	applyNameSplit(fields)
	c, err := s.recruit.CreateCandidate(tenantID, CandidateInput{
		FirstName: fields["first_name"], LastName: fields["last_name"],
		Email: fields["email"], Phone: fields["phone"],
		Address: fields["address"], DateOfBirth: fields["date_of_birth"],
		Education: fields["education"], Experience: fields["experience"],
		Skills: fields["skills"], Source: "Google Forms",
	}, actorID, ip, ua)
	if err != nil {
		return nil, err
	}
	_, err = s.recruit.UpdateCandidate(tenantID, c.ID, CandidateInput{
		Email:   "", // never change via sync
		Address: fields["address"], DateOfBirth: fields["date_of_birth"],
		Education: fields["education"], Experience: fields["experience"],
		Skills:     fields["skills"],
		ResumePath: fields["resume_path"],
		Notes:      fields["notes"],
	}, actorID, ip, ua)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *GoogleFormService) SyncStatus(tenantID, jobID string) (*models.GoogleFormIntegration, dto.GoogleSyncCounters, error) {
	integ, err := s.integrations.GetByJob(tenantID, jobID)
	if err != nil {
		return nil, dto.GoogleSyncCounters{}, ErrNotFound
	}
	total, imported, duplicate, failed, err := s.responses.Counts(tenantID, integ.ID)
	if err != nil {
		return nil, dto.GoogleSyncCounters{}, err
	}
	return integ, dto.GoogleSyncCounters{Total: total, Imported: imported, Duplicates: duplicate, Failed: failed}, nil
}

func (s *GoogleFormService) Responses(tenantID, jobID, status string, p utils.Pagination) ([]models.GoogleFormResponse, int64, error) {
	integ, err := s.integrations.GetByJob(tenantID, jobID)
	if err != nil {
		return nil, 0, ErrNotFound
	}
	return s.responses.List(tenantID, integ.ID, status, p.Offset, p.Limit)
}

// --- helpers ---

func (s *GoogleFormService) failSync(tenantID string, integ *models.GoogleFormIntegration, err error) error {
	wrapped := mapGoogleAPIError(err)
	_ = s.integrations.Update(tenantID, integ.ID, map[string]interface{}{
		"sync_error": wrapped.Error(), "status": models.GoogleFormStatusError,
	})
	return wrapped
}

func mapGoogleAPIError(err error) error {
	switch {
	case errors.Is(err, google.ErrPermissionDenied):
		return fmt.Errorf("%w: %v", ErrGooglePermissionDenied, err)
	case errors.Is(err, google.ErrInvalidSpreadsheet):
		return fmt.Errorf("%w: %v", ErrGoogleInvalidSpreadsheet, err)
	case errors.Is(err, google.ErrRateLimit):
		return ErrGoogleRateLimit
	case errors.Is(err, google.ErrAuthExpired), errors.Is(err, google.ErrNotAuthorized):
		return fmt.Errorf("%w: %v", ErrGoogleNotAuthorized, err)
	case errors.Is(err, google.ErrNetwork):
		return ErrGoogleNetwork
	case errors.Is(err, google.ErrAPIStatus):
		return fmt.Errorf("%w: %v", ErrGoogleAPIStatus, err)
	case errors.Is(err, google.ErrNotConfigured):
		return ErrGoogleNotConfigured
	default:
		return err
	}
}

func (s *GoogleFormService) storedRules(raw datatypes.JSON) ([]FieldMap, bool, error) {
	if len(raw) == 0 {
		return defaultFieldRules, false, nil
	}
	var rules []FieldMap
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, false, fmt.Errorf("invalid field mapping: %w", err)
	}
	if len(rules) == 0 {
		return defaultFieldRules, false, nil
	}
	return rules, true, nil
}

func (s *GoogleFormService) recordResponse(tenantID string, integ *models.GoogleFormIntegration, extID string, candidateID, appID *string, submitted *time.Time, fields map[string]string, status, msg string) error {
	raw, _ := json.Marshal(fields)
	r := &models.GoogleFormResponse{
		TenantID:           tenantID,
		IntegrationID:      integ.ID,
		ExternalResponseID: extID,
		CandidateID:        candidateID,
		ApplicationID:      appID,
		RawResponse:        datatypes.JSON(raw),
		SubmittedAt:        submitted,
		Status:             status,
		ErrorMessage:       msg,
	}
	if status == models.GoogleFormResponseImported {
		now := time.Now()
		r.ImportedAt = &now
	}
	return s.responses.Create(r)
}

func validateFieldMapping(rules []FieldMap) error {
	if len(rules) == 0 {
		return nil
	}
	for _, r := range rules {
		if !allowedTargets[r.Target] {
			return fmt.Errorf("%w: %q", ErrGoogleTargetInvalid, r.Target)
		}
		if strings.TrimSpace(r.Source) == "" {
			return fmt.Errorf("google field mapping source cannot be empty")
		}
	}
	return nil
}

func marshalFieldMapping(rules []FieldMap) datatypes.JSON {
	if len(rules) == 0 {
		return nil
	}
	b, err := json.Marshal(rules)
	if err != nil {
		return nil
	}
	return datatypes.JSON(b)
}

func resolveFieldMapping(headers []string, rules []FieldMap, strict bool) (map[string]int, error) {
	idx := headerIndexMap(headers)
	out := map[string]int{}
	for _, r := range rules {
		if !allowedTargets[r.Target] && strict {
			return nil, fmt.Errorf("%w: %q", ErrGoogleTargetInvalid, r.Target)
		}
		col, ok := idx[normalizeHeader(r.Source)]
		if !ok {
			if strict {
				return nil, fmt.Errorf("%w: %q", ErrGoogleMissingHeader, r.Source)
			}
			continue
		}
		if _, exists := out[r.Target]; !exists {
			out[r.Target] = col
		}
	}
	return out, nil
}

func headerIndexMap(headers []string) map[string]int {
	out := make(map[string]int, len(headers))
	for i, h := range headers {
		key := normalizeHeader(h)
		if key == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = i
		}
	}
	return out
}

func normalizeHeader(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func valuesFromRow(headerIdx map[string]int, row []string) map[string]string {
	out := make(map[string]string, len(headerIdx))
	for target, col := range headerIdx {
		if col < len(row) {
			v := strings.TrimSpace(row[col])
			if v != "" {
				out[target] = v
			}
		}
	}
	return out
}

func applyNameSplit(fields map[string]string) {
	first := fields["first_name"]
	if first == "" {
		return
	}
	if fields["last_name"] != "" {
		return
	}
	parts := strings.Fields(first)
	if len(parts) > 1 {
		fields["first_name"] = strings.Join(parts[:len(parts)-1], " ")
		fields["last_name"] = parts[len(parts)-1]
	}
}

func externalResponseID(integ *models.GoogleFormIntegration, rowIdx int, explicit string) string {
	if explicit != "" {
		return "gsr:" + explicit
	}
	return fmt.Sprintf("gsrow:%s:%d", integ.ID, rowIdx)
}

func isBlankRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func validEmail(s string) bool {
	if len(s) < 3 || !strings.Contains(s, "@") {
		return false
	}
	at := strings.Index(s, "@")
	if at == 0 || at == len(s)-1 || strings.ContainsAny(s, " \t\n") {
		return false
	}
	return strings.Contains(s[at+1:], ".")
}

func parseSubmittedAt(s string) *time.Time {
	if s == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"1/2/2006 15:04:05",
		"01/02/2006 15:04",
		"1/2/2006",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return &t
		}
	}
	return nil
}

func appliedDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		now := time.Now()
		t = &now
	}
	return t.Format("2006-01-02")
}

func isHTTPURL(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")
}

func mapGoogleOAuthError(err error) error {
	switch {
	case errors.Is(err, google.ErrInvalidState):
		return ErrGoogleOAuthStateInvalid
	case errors.Is(err, google.ErrNotConfigured):
		return ErrGoogleNotConfigured
	case errors.Is(err, google.ErrNotAuthorized):
		return ErrGoogleNotAuthorized
	case errors.Is(err, google.ErrCodeExchange):
		return fmt.Errorf("%w: %v", ErrGoogleNotAuthorized, err)
	default:
		return err
	}
}
