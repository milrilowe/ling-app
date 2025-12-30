package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"ling-app/api/internal/models"
)

// phonemeStatsRepository implements PhonemeStatsRepository using GORM.
type phonemeStatsRepository struct{}

// NewPhonemeStatsRepository creates a new GORM-backed phoneme stats repository.
func NewPhonemeStatsRepository() PhonemeStatsRepository {
	return &phonemeStatsRepository{}
}

func (r *phonemeStatsRepository) Upsert(exec Executor, stats *models.PhonemeStats) error {
	return exec.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "phoneme"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_attempts": clause.Expr{SQL: "phoneme_stats.total_attempts + ?", Vars: []interface{}{stats.TotalAttempts}},
			"correct_count":  clause.Expr{SQL: "phoneme_stats.correct_count + ?", Vars: []interface{}{stats.CorrectCount}},
			"deletion_count": clause.Expr{SQL: "phoneme_stats.deletion_count + ?", Vars: []interface{}{stats.DeletionCount}},
			"updated_at":     clause.Expr{SQL: "NOW()"},
		}),
	}).Create(stats).Error
}

func (r *phonemeStatsRepository) FindByUserID(exec Executor, userID uuid.UUID) ([]models.PhonemeStats, error) {
	var stats []models.PhonemeStats
	err := exec.Where("user_id = ?", userID).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *phonemeStatsRepository) GetAccuracyRanking(exec Executor, userID uuid.UUID) ([]PhonemeAccuracy, error) {
	var phonemeStats []PhonemeAccuracy
	err := exec.Model(&models.PhonemeStats{}).
		Select("phoneme, total_attempts, correct_count, deletion_count, (CAST(correct_count AS FLOAT) / CAST(total_attempts AS FLOAT) * 100) as accuracy").
		Where("user_id = ?", userID).
		Order("accuracy ASC").
		Scan(&phonemeStats).Error
	if err != nil {
		return nil, err
	}
	return phonemeStats, nil
}

// phonemeSubstitutionRepository implements PhonemeSubstitutionRepository using GORM.
type phonemeSubstitutionRepository struct{}

// NewPhonemeSubstitutionRepository creates a new GORM-backed phoneme substitution repository.
func NewPhonemeSubstitutionRepository() PhonemeSubstitutionRepository {
	return &phonemeSubstitutionRepository{}
}

func (r *phonemeSubstitutionRepository) Upsert(exec Executor, sub *models.PhonemeSubstitution) error {
	return exec.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "expected_phoneme"}, {Name: "actual_phoneme"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"occurrence_count": clause.Expr{SQL: "phoneme_substitutions.occurrence_count + ?", Vars: []interface{}{sub.OccurrenceCount}},
			"updated_at":       clause.Expr{SQL: "NOW()"},
		}),
	}).Create(sub).Error
}

func (r *phonemeSubstitutionRepository) FindTopByUserID(exec Executor, userID uuid.UUID, limit int) ([]models.PhonemeSubstitution, error) {
	var substitutions []models.PhonemeSubstitution
	err := exec.Where("user_id = ?", userID).
		Order("occurrence_count DESC").
		Limit(limit).
		Find(&substitutions).Error
	if err != nil {
		return nil, err
	}
	return substitutions, nil
}
