package platform

import (
	"fmt"

	"gorm.io/gorm"
)

// ValidateDeptClosure rebuilds the sys_dept_closure table from scratch
// based on the current sys_dept.ancestors strings. Safe to call at
// startup; clears and re-inserts closure rows. Operates on all tenants
// (platform + shops) since the closure table is global.
func ValidateDeptClosure(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("TRUNCATE TABLE sys_dept_closure").Error; err != nil {
			return fmt.Errorf("truncate closure: %w", err)
		}

		if err := tx.Exec(`
			INSERT INTO sys_dept_closure (tenant_id, ancestor_id, descendant_id, depth)
			SELECT tenant_id, id, id, 0 FROM sys_dept
		`).Error; err != nil {
			return fmt.Errorf("insert self-rows: %w", err)
		}

		if err := tx.Exec(`
			INSERT INTO sys_dept_closure (tenant_id, ancestor_id, descendant_id, depth)
			SELECT d.tenant_id, CAST(trim(a) AS BIGINT), d.id,
			       array_length(string_to_array(d.ancestors, ','), 1) - ord + 1
			FROM sys_dept d
			CROSS JOIN LATERAL unnest(string_to_array(d.ancestors, ','))
				WITH ORDINALITY AS x(a, ord)
			WHERE trim(a) <> '0' AND trim(a) <> ''
		`).Error; err != nil {
			return fmt.Errorf("insert ancestor rows: %w", err)
		}

		return nil
	})
}
