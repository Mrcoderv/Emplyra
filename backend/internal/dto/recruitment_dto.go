package dto

type JobPostRequest struct {
	Title        string `json:"title" binding:"required,min=2,max=200"`
	DepartmentID string `json:"department_id"`
	Description  string `json:"description" binding:"max=4000"`
	Requirements string `json:"requirements" binding:"max=4000"`
	Vacancies    int    `json:"vacancies"`
	Status       string `json:"status"`
	Deadline     string `json:"deadline"`
}

type CandidateRequest struct {
	FirstName   string `json:"first_name" binding:"required,max=100"`
	LastName    string `json:"last_name" binding:"max=100"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone" binding:"max=30"`
	Source      string `json:"source" binding:"max=100"`
	Status      string `json:"status"`
	Notes       string `json:"notes" binding:"max=2000"`
	ResumePath  string `json:"resume_path" binding:"max=500"`
	Address     string `json:"address" binding:"max=500"`
	DateOfBirth string `json:"date_of_birth" binding:"max=40"`
	Education   string `json:"education" binding:"max=500"`
	Experience  string `json:"experience" binding:"max=500"`
	Skills      string `json:"skills" binding:"max=1000"`
}

type ApplicationRequest struct {
	JobPostID   string `json:"job_post_id" binding:"required"`
	CandidateID string `json:"candidate_id" binding:"required"`
	AppliedDate string `json:"applied_date"`
	CoverLetter string `json:"cover_letter" binding:"max=4000"`
}

type ApplicationStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Note   string `json:"note" binding:"max=2000"`
}

type InterviewRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	InterviewerID string `json:"interviewer_id"`
	ScheduledAt   string `json:"scheduled_at" binding:"required"`
	DurationMin   int    `json:"duration_minutes"`
	Type          string `json:"type"`
}

type InterviewFeedbackRequest struct {
	Status   string   `json:"status"`
	Feedback string   `json:"feedback" binding:"max=4000"`
	Score    *float64 `json:"score"`
}

type OnboardingRequest struct {
	EmployeeID  string   `json:"employee_id" binding:"required"`
	CandidateID string   `json:"candidate_id"`
	StartDate   string   `json:"start_date" binding:"required"`
	Tasks       []string `json:"tasks"`
	Notes       string   `json:"notes" binding:"max=2000"`
}

type OnboardingUpdateRequest struct {
	Status string   `json:"status"`
	Tasks  []string `json:"tasks"`
	Notes  string   `json:"notes" binding:"max=2000"`
}

type HireCandidateRequest struct {
	JobPostID     string `json:"job_post_id" binding:"required"`
	ApplicationID string `json:"application_id" binding:"required"`
	EmployeeCode  string `json:"employee_code"`
	DepartmentID  string `json:"department_id"`
	DesignationID string `json:"designation_id"`
	ManagerID     string `json:"manager_id"`
	JoiningDate   string `json:"joining_date"`
	CreateUser    bool   `json:"create_user"`
}
