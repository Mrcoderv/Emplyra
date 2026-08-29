package google

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SheetsReader interface {
	ListSheets(spreadsheetID string) ([]string, error)
	GetValues(spreadsheetID, sheetName string) ([][]string, error)
}

type SheetsClient struct {
	tm   *TokenManager
	http *http.Client
}

func NewSheetsClient(tm *TokenManager) *SheetsClient {
	return &SheetsClient{
		tm:   tm,
		http: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *SheetsClient) ListSheets(spreadsheetID string) ([]string, error) {
	endpoint := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s", url.PathEscape(spreadsheetID))
	var out struct {
		Sheets []struct {
			Properties struct {
				Title string `json:"title"`
			} `json:"properties"`
		} `json:"sheets"`
	}
	if err := s.getJSON(endpoint, &out); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(out.Sheets))
	for _, sh := range out.Sheets {
		names = append(names, sh.Properties.Title)
	}
	return names, nil
}

func (s *SheetsClient) GetValues(spreadsheetID, sheetName string) ([][]string, error) {
	if sheetName == "" {
		return nil, fmt.Errorf("sheet name is required")
	}
	endpoint := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?valueRenderOption=UNFORMATTED_VALUE",
		url.PathEscape(spreadsheetID), url.PathEscape(sheetName),
	)
	var out struct {
		Values [][]interface{} `json:"values"`
	}
	if err := s.getJSON(endpoint, &out); err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(out.Values))
	for _, cells := range out.Values {
		row := make([]string, 0, len(cells))
		for _, c := range cells {
			row = append(row, cellString(c))
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func cellString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "TRUE"
		}
		return "FALSE"
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (s *SheetsClient) getJSON(endpoint string, out interface{}) error {
	token, err := s.tm.AccessToken()
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.http.Do(req)
	if err != nil {
		return ErrNetwork
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mapAPIError(resp)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrNetwork
	}
	return json.Unmarshal(body, out)
}

func mapAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	msg := strings.TrimSpace(string(body))
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrAuthExpired
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrPermissionDenied, snippet(msg))
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrInvalidSpreadsheet, snippet(msg))
	case http.StatusTooManyRequests:
		return ErrRateLimit
	default:
		return fmt.Errorf("%w (status %d): %s", ErrAPIStatus, resp.StatusCode, snippet(msg))
	}
}

func snippet(s string) string {
	if len(s) > 200 {
		return s[:200]
	}
	return s
}
