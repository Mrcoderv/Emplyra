package dto

type GoalRequest struct {
	EmployeeID  string `json:"employee_id"`
	Title       string `json:"title" binding:"required,min=2,max=200"`
	Description string `json:"description" binding:"max=2000"`
	TargetDate  string `json:"target_date"`
	Weight      int    `json:"weight"`
	Status      string `json:"status"`
}

type KPIRequest struct {
	EmployeeID  string   `json:"employee_id"`
	Name        string   `json:"name" binding:"required,min=2,max=200"`
	Description string   `json:"description" binding:"max=1000"`
	Target      string   `json:"target"`
	Actual      string   `json:"actual"`
	Unit        string   `json:"unit"`
	Weight      int      `json:"weight"`
	Period      string   `json:"period"`
	Score       *float64 `json:"score"`
}

type ReviewRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	ReviewerID string `json:"reviewer_id"`
	Period     string `json:"period" binding:"required,max=50"`
	DueDate    string `json:"due_date"`
}

type ReviewSubmissionRequest struct {
	SelfEvaluation  string   `json:"self_evaluation" binding:"max=4000"`
	ManagerFeedback string   `json:"manager_feedback" binding:"max=4000"`
	Score           *float64 `json:"score"`
	Status          string   `json:"status"`
}

type TrainingProgramRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=200"`
	Description string `json:"description" binding:"max=4000"`
	Provider    string `json:"provider" binding:"max=200"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	Location    string `json:"location" binding:"max=200"`
	MaxSeats    int    `json:"max_seats"`
	Status      string `json:"status"`
}

type TrainingScheduleRequest struct {
	ProgramID string `json:"program_id" binding:"required"`
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"start_time" binding:"max=10"`
	EndTime   string `json:"end_time" binding:"max=10"`
	Trainer   string `json:"trainer" binding:"max=200"`
	Location  string `json:"location" binding:"max=200"`
	MaxSeats  int    `json:"max_seats"`
}

type EnrollmentRequest struct {
	ProgramID  string `json:"program_id" binding:"required"`
	EmployeeID string `json:"employee_id"`
}

type EnrollmentUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}
