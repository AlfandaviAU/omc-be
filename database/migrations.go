package database

import (
	"log"
	"time"

	"github.com/davi/omc-be/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// schemaMigration records data migrations that already ran.
// Schema changes are handled by AutoMigrate.
type schemaMigration struct {
	ID        string `gorm:"primaryKey"`
	AppliedAt time.Time
}

// Append new migrations to the end. Never edit one that has shipped.
var migrations = []struct {
	id  string
	run func(tx *gorm.DB) error
}{
	{"2026_11_08_relink_shotgun_attachments", relinkShotgunAttachments},
}

func runMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return err
	}

	for _, m := range migrations {
		err := db.Transaction(func(tx *gorm.DB) error {
			var applied int64
			if err := tx.Model(&schemaMigration{}).Where("id = ?", m.id).Count(&applied).Error; err != nil {
				return err
			}
			if applied > 0 {
				return nil
			}

			if err := m.run(tx); err != nil {
				return err
			}
			log.Println("Applied migration", m.id)
			// DoNothing lets instances booting at the same time both succeed.
			return tx.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&schemaMigration{ID: m.id, AppliedAt: time.Now()}).Error
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// relinkShotgunAttachments points the seeded "Shotgun" links at "Pump Shotgun".
// No product is named "Shotgun", so the store showed a kit with no gun in it.
func relinkShotgunAttachments(tx *gorm.DB) error {
	var oldProducts, newProducts int64
	if err := tx.Model(&models.Product{}).Where("LOWER(name) = ?", "shotgun").Count(&oldProducts).Error; err != nil {
		return err
	}
	if err := tx.Model(&models.Product{}).Where("LOWER(name) = ?", "pump shotgun").Count(&newProducts).Error; err != nil {
		return err
	}
	if oldProducts > 0 || newProducts == 0 {
		return nil
	}

	return tx.Model(&models.WeaponAttachment{}).
		Where("weapon_name = ?", "Shotgun").
		Update("weapon_name", "Pump Shotgun").Error
}
