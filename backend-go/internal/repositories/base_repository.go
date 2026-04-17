package repositories

import (
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// baseRepository — shared GORM helpers used by every concrete repository.
// Never instantiated directly; embedded by each domain repository.
// ---------------------------------------------------------------------------

type baseRepository struct {
	db *gorm.DB
}

// paginate returns a *gorm.DB scoped to the requested page.
// page is 1-based. pageSize defaults to 20 if <= 0.
func (r *baseRepository) paginate(page, pageSize int) *gorm.DB {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return r.db.Limit(pageSize).Offset(offset)
}

// applyActiveFilter scopes a query to is_active = true.
func applyActiveFilter(db *gorm.DB) *gorm.DB {
	return db.Where("is_active = ?", true)
}

// applyStatusFilter scopes a query to a specific status column value.
func applyStatusFilter(db *gorm.DB, status string) *gorm.DB {
	return db.Where("status = ?", status)
}

// countRows returns total row count for a model with the given conditions.
func countRows(db *gorm.DB, model interface{}) (int64, error) {
	var count int64
	err := db.Model(model).Count(&count).Error
	return count, err
}
