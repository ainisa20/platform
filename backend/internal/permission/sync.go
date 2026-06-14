package permission

import (
	"time"

	"platform/internal/model/entity"
	"platform/internal/model/enum"

	"gorm.io/gorm"
)

func SyncPermissions(db *gorm.DB, manifests ...Manifest) error {
	for _, m := range manifests {
		if err := syncNodes(db, m.Nodes, 0, m.SystemType); err != nil {
			return err
		}
		if err := pruneObsolete(db, m); err != nil {
			return err
		}
	}
	return nil
}

func pruneObsolete(db *gorm.DB, m Manifest) error {
	keep := map[string]struct{}{}
	for _, n := range m.Nodes {
		keep[n.Name] = struct{}{}
		for _, b := range n.Buttons {
			keep[b.Code] = struct{}{}
		}
		for _, c := range n.Children {
			keep[c.Name] = struct{}{}
			for _, b := range c.Buttons {
				keep[b.Code] = struct{}{}
			}
		}
	}

	var stale []entity.SysPermission
	if err := db.Where("system_type = ? AND auto_synced = ?", m.SystemType, true).
		Find(&stale).Error; err != nil {
		return err
	}
	for _, p := range stale {
		key := p.PermsCode
		if key == "" {
			key = p.Name
		}
		if _, ok := keep[key]; ok {
			continue
		}
		if err := db.Where("permission_id = ?", p.ID).Delete(&entity.SysRolePermission{}).Error; err != nil {
			return err
		}
		if err := db.Delete(&entity.SysPermission{}, p.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func syncNodes(db *gorm.DB, nodes []Node, parentID uint64, systemType string) error {
	for i := range nodes {
		n := &nodes[i]

		id, err := upsertNode(db, entity.SysPermission{
			ParentID:   parentID,
			SystemType: systemType,
			Name:       n.Name,
			Type:       n.Type,
			Path:       n.Path,
			Component:  n.Component,
			PermsCode:  "",
			Icon:       n.Icon,
			Sort:       int16(n.Sort),
			Visible:    true,
			Status:     enum.StatusEnabled,
			AutoSynced: true,
			UpdatedAt:  time.Now(),
		})
		if err != nil {
			return err
		}

		for j := range n.Buttons {
			b := &n.Buttons[j]
			if _, err := upsertNode(db, entity.SysPermission{
				ParentID:   id,
				SystemType: systemType,
				Name:       b.Name,
				Type:       enum.PermTypeButton,
				PermsCode:  b.Code,
				Sort:       int16(j + 1),
				Visible:    true,
				Status:     enum.StatusEnabled,
				AutoSynced: true,
				UpdatedAt:  time.Now(),
			}); err != nil {
				return err
			}
		}

		if len(n.Children) > 0 {
			if err := syncNodes(db, n.Children, id, systemType); err != nil {
				return err
			}
		}
	}
	return nil
}

func upsertNode(db *gorm.DB, rec entity.SysPermission) (uint64, error) {
	var existing entity.SysPermission

	if rec.PermsCode == "" {
		result := db.Where("system_type = ? AND parent_id = ? AND type = ? AND name = ? AND perms_code = ''",
			rec.SystemType, rec.ParentID, rec.Type, rec.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&rec).Error; err != nil {
				return 0, err
			}
			return rec.ID, nil
		}
		if result.Error != nil {
			return 0, result.Error
		}
	} else {
		result := db.Where("system_type = ? AND perms_code = ?",
			rec.SystemType, rec.PermsCode).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&rec).Error; err != nil {
				return 0, err
			}
			return rec.ID, nil
		}
		if result.Error != nil {
			return 0, result.Error
		}
	}

	if err := db.Model(&existing).Updates(map[string]interface{}{
		"name":       rec.Name,
		"parent_id":  rec.ParentID,
		"path":       rec.Path,
		"component":  rec.Component,
		"perms_code": rec.PermsCode,
		"icon":       rec.Icon,
		"sort":       rec.Sort,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return 0, err
	}
	return existing.ID, nil
}
