package dto

type GoogleFieldMap struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type GoogleFormRequest struct {
	GoogleFormURL string           `json:"google_form_url" binding:"omitempty,url,max=500"`
	GoogleSheetID string           `json:"google_sheet_id" binding:"max=200"`
	ResponseSheet string           `json:"response_sheet_name" binding:"max=200"`
	HeaderRow     int              `json:"header_row"`
	Status        string           `json:"status"`
	StatusDetail  string           `json:"status_detail" binding:"max=255"`
	FieldMapping  []GoogleFieldMap `json:"field_mapping"`
}

type GoogleFormSyncRequest struct {
	Mode string `json:"mode"` // "incremental" (default) | "full"
}

type GoogleFormsResponseQuery struct {
	Page     string `form:"page"`
	PageSize string `form:"page_size"`
	Status   string `form:"status"`
}

type GoogleOAuthAuthorizeResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

type GoogleSyncResult struct {
	Imported   int `json:"imported"`
	Duplicates int `json:"duplicates"`
	Failed     int `json:"failed"`
	TotalRows  int `json:"total_rows"`
}

type GoogleSyncCounters struct {
	Total      int64 `json:"total"`
	Imported   int64 `json:"imported"`
	Duplicates int64 `json:"duplicates"`
	Failed     int64 `json:"failed"`
}
