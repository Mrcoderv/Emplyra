package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

type GoogleFormIntegrationRepository struct {
	db *gorm.DB
}

func NewGoogleFormIntegrationRepository(db *gorm.DB) *GoogleFormIntegrationRepository {
	return &GoogleFormIntegrationRepository{db: db}
}

func (r *GoogleFormIntegrationRepository) Create(i *models.GoogleFormIntegration) error {
	return r.db.Create(i).Error
}

func (r *GoogleFormIntegrationRepository) GetByJob(tenantID, jobID string) (*models.GoogleFormIntegration, error) {
	var i models.GoogleFormIntegration
	err := r.db.Scopes(TenantScope(tenantID)).Preload("JobPost").First(&i, "job_id = ?", jobID).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *GoogleFormIntegrationRepository) FindByID(tenantID, id string) (*models.GoogleFormIntegration, error) {
	var i models.GoogleFormIntegration
	err := r.db.Scopes(TenantScope(tenantID)).Preload("JobPost").First(&i, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *GoogleFormIntegrationRepository) Update(tenantID, id string, fields map[string]interface{}) error {
	return r.db.Model(&models.GoogleFormIntegration{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(fields).Error
}

func (r *GoogleFormIntegrationRepository) DeleteByJob(tenantID, jobID string) error {
	return r.db.Scopes(TenantScope(tenantID)).Where("job_id = ?", jobID).Delete(&models.GoogleFormIntegration{}).Error
}

type GoogleFormResponseRepository struct {
	db *gorm.DB
}

func NewGoogleFormResponseRepository(db *gorm.DB) *GoogleFormResponseRepository {
	return &GoogleFormResponseRepository{db: db}
}

func (r *GoogleFormResponseRepository) Create(resp *models.GoogleFormResponse) error {
	return r.db.Create(resp).Error
}

func (r *GoogleFormResponseRepository) ExistsByExternalID(tenantID, externalID string) (bool, error) {
	var n int64
	err := r.db.Scopes(TenantScope(tenantID)).Model(&models.GoogleFormResponse{}).Where("external_response_id = ?", externalID).Count(&n).Error
	return n > 0, err
}

func (r *GoogleFormResponseRepository) List(tenantID, integrationID, status string, offset, limit int) ([]models.GoogleFormResponse, int64, error) {
	var items []models.GoogleFormResponse
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.GoogleFormResponse{}).Where("integration_id = ?", integrationID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Candidate").
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *GoogleFormResponseRepository) Counts(tenantID, integrationID string) (total, imported, duplicate, failed int64, err error) {
	base := func() *gorm.DB {
		return r.db.Scopes(TenantScope(tenantID)).Model(&models.GoogleFormResponse{}).Where("integration_id = ?", integrationID)
	}
	if err = base().Count(&total).Error; err != nil {
		return
	}
	if err = base().Where("status = ?", models.GoogleFormResponseImported).Count(&imported).Error; err != nil {
		return
	}
	if err = base().Where("status = ?", models.GoogleFormResponseDuplicate).Count(&duplicate).Error; err != nil {
		return
	}
	err = base().Where("status = ?", models.GoogleFormResponseError).Count(&failed).Error
	return
}

type GoogleOAuthTokenRepository struct {
	db *gorm.DB
}

func NewGoogleOAuthTokenRepository(db *gorm.DB) *GoogleOAuthTokenRepository {
	return &GoogleOAuthTokenRepository{db: db}
}

func (r *GoogleOAuthTokenRepository) Get(key string) (string, error) {
	var t models.GoogleOAuthToken
	err := r.db.Where("key = ?", key).First(&t).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
		_ = r.Delete(key)
		return "", nil
	}
	return t.Data, nil
}

func (r *GoogleOAuthTokenRepository) Set(key, data string, expiresAt *time.Time) error {
	var t models.GoogleOAuthToken
	err := r.db.Where("key = ?", key).First(&t).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&models.GoogleOAuthToken{Key: key, Data: data, ExpiresAt: expiresAt}).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&models.GoogleOAuthToken{}).Where("key = ?", key).
		Updates(map[string]interface{}{"data": data, "expires_at": expiresAt}).Error
}

func (r *GoogleOAuthTokenRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&models.GoogleOAuthToken{}).Error
}
