package repositories

import "gorm.io/gorm"

// TenantScope appends a tenant_id predicate when id is non-empty. Pass an empty
// id to query platform-wide (used by platform operators / support mode).
func TenantScope(id string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if id != "" {
			db = db.Where("tenant_id = ?", id)
		}
		return db
	}
}
