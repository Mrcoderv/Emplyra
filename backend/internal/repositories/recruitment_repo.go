package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type JobPostRepository struct {
	db *gorm.DB
}

func NewJobPostRepository(db *gorm.DB) *JobPostRepository {
	return &JobPostRepository{db: db}
}

func (r *JobPostRepository) Create(j *models.JobPost) error { return r.db.Create(j).Error }

func (r *JobPostRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.JobPost{}).Where("id = ?", id).Updates(fields).Error
}

func (r *JobPostRepository) Delete(id string) error {
	return r.db.Delete(&models.JobPost{}, "id = ?", id).Error
}

func (r *JobPostRepository) FindByID(id string) (*models.JobPost, error) {
	var j models.JobPost
	err := r.db.Preload("Department").Preload("PostedByUser").First(&j, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *JobPostRepository) List(p utils.Pagination, departmentID, status string, search string) ([]models.JobPost, int64, error) {
	var items []models.JobPost
	var total int64
	q := r.db.Model(&models.JobPost{})
	if departmentID != "" {
		q = q.Where("department_id = ?", departmentID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Department").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

type CandidateRepository struct {
	db *gorm.DB
}

func NewCandidateRepository(db *gorm.DB) *CandidateRepository {
	return &CandidateRepository{db: db}
}

func (r *CandidateRepository) Create(c *models.Candidate) error { return r.db.Create(c).Error }

func (r *CandidateRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Candidate{}).Where("id = ?", id).Updates(fields).Error
}

func (r *CandidateRepository) Delete(id string) error {
	return r.db.Delete(&models.Candidate{}, "id = ?", id).Error
}

func (r *CandidateRepository) FindByID(id string) (*models.Candidate, error) {
	var c models.Candidate
	err := r.db.First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CandidateRepository) FindByEmail(email string) (*models.Candidate, error) {
	var c models.Candidate
	err := r.db.Where("LOWER(email) = LOWER(?)", email).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CandidateRepository) List(p utils.Pagination, status, search string) ([]models.Candidate, int64, error) {
	var items []models.Candidate
	var total int64
	q := r.db.Model(&models.Candidate{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

func (r *CandidateRepository) AlreadyEmployeeEmail(email string) bool {
	var n int64
	r.db.Model(&models.Employee{}).Where("LOWER(email) = LOWER(?)", email).Count(&n)
	return n > 0
}

type ApplicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(a *models.Application) error { return r.db.Create(a).Error }

func (r *ApplicationRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Application{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ApplicationRepository) FindByID(id string) (*models.Application, error) {
	var a models.Application
	err := r.db.Preload("JobPost").Preload("Candidate").Preload("Reviewer").First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ApplicationRepository) Exists(candidateID, jobPostID string) (bool, error) {
	var n int64
	err := r.db.Model(&models.Application{}).
		Where("candidate_id = ? AND job_post_id = ?", candidateID, jobPostID).Count(&n).Error
	return n > 0, err
}

func (r *ApplicationRepository) List(p utils.Pagination, jobID, candidateID, status string) ([]models.Application, int64, error) {
	var items []models.Application
	var total int64
	q := r.db.Model(&models.Application{})
	if jobID != "" {
		q = q.Where("job_post_id = ?", jobID)
	}
	if candidateID != "" {
		q = q.Where("candidate_id = ?", candidateID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("JobPost").Preload("Candidate").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

type InterviewRepository struct {
	db *gorm.DB
}

func NewInterviewRepository(db *gorm.DB) *InterviewRepository {
	return &InterviewRepository{db: db}
}

func (r *InterviewRepository) Create(i *models.Interview) error { return r.db.Create(i).Error }

func (r *InterviewRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Interview{}).Where("id = ?", id).Updates(fields).Error
}

func (r *InterviewRepository) FindByID(id string) (*models.Interview, error) {
	var i models.Interview
	err := r.db.Preload("Application.Candidate").Preload("Application.JobPost").Preload("Interviewer").First(&i, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *InterviewRepository) List(p utils.Pagination, applicationID, interviewerID, status string, from, to *time.Time) ([]models.Interview, int64, error) {
	var items []models.Interview
	var total int64
	q := r.db.Model(&models.Interview{})
	if applicationID != "" {
		q = q.Where("application_id = ?", applicationID)
	}
	if interviewerID != "" {
		q = q.Where("interviewer_id = ?", interviewerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if from != nil {
		q = q.Where("scheduled_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("scheduled_at <= ?", *to)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Application.Candidate").Preload("Interviewer").Order("scheduled_at ASC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

type OnboardingRepository struct {
	db *gorm.DB
}

func NewOnboardingRepository(db *gorm.DB) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

func (r *OnboardingRepository) Create(o *models.Onboarding) error { return r.db.Create(o).Error }

func (r *OnboardingRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Onboarding{}).Where("id = ?", id).Updates(fields).Error
}

func (r *OnboardingRepository) FindByID(id string) (*models.Onboarding, error) {
	var o models.Onboarding
	err := r.db.Preload("Employee").Preload("Candidate").First(&o, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OnboardingRepository) List(p utils.Pagination, employeeID, status string) ([]models.Onboarding, int64, error) {
	var items []models.Onboarding
	var total int64
	q := r.db.Model(&models.Onboarding{})
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}
