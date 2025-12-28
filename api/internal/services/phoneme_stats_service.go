package services

import (
	"ling-app/api/internal/db"
	"ling-app/api/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// PhonemeStatsService handles phoneme statistics aggregation
type PhonemeStatsService struct {
	DB *db.DB
}

// NewPhonemeStatsService creates a new phoneme stats service
func NewPhonemeStatsService(database *db.DB) *PhonemeStatsService {
	return &PhonemeStatsService{DB: database}
}

// RecordPhonemeResults processes phoneme details from pronunciation analysis
// and updates the user's aggregate statistics
func (s *PhonemeStatsService) RecordPhonemeResults(userID uuid.UUID, phonemeDetails []PhonemeDetail) error {
	if len(phonemeDetails) == 0 {
		return nil
	}

	// Aggregate stats from this analysis
	statsMap := make(map[string]*models.PhonemeStats)
	subsMap := make(map[string]*models.PhonemeSubstitution)

	for _, detail := range phonemeDetails {
		// Skip insertions (extra phonemes user added) - we only track expected phonemes
		if detail.Type == "insert" {
			continue
		}

		expected := detail.Expected
		if expected == "" {
			continue
		}

		// Initialize or update stats for this phoneme
		if _, exists := statsMap[expected]; !exists {
			statsMap[expected] = &models.PhonemeStats{
				UserID:        userID,
				Phoneme:       expected,
				TotalAttempts: 0,
				CorrectCount:  0,
				DeletionCount: 0,
			}
		}
		statsMap[expected].TotalAttempts++

		if detail.Type == "match" {
			statsMap[expected].CorrectCount++
		} else if detail.Type == "delete" {
			statsMap[expected].DeletionCount++
		} else if detail.Type == "substitute" && detail.Actual != "" {
			// Track substitution pattern
			subKey := expected + "->" + detail.Actual
			if _, exists := subsMap[subKey]; !exists {
				subsMap[subKey] = &models.PhonemeSubstitution{
					UserID:          userID,
					ExpectedPhoneme: expected,
					ActualPhoneme:   detail.Actual,
					OccurrenceCount: 0,
				}
			}
			subsMap[subKey].OccurrenceCount++
		}
	}

	// Upsert phoneme stats
	for _, stats := range statsMap {
		err := s.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "phoneme"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_attempts": clause.Expr{SQL: "phoneme_stats.total_attempts + ?", Vars: []interface{}{stats.TotalAttempts}},
				"correct_count":  clause.Expr{SQL: "phoneme_stats.correct_count + ?", Vars: []interface{}{stats.CorrectCount}},
				"deletion_count": clause.Expr{SQL: "phoneme_stats.deletion_count + ?", Vars: []interface{}{stats.DeletionCount}},
				"updated_at":     clause.Expr{SQL: "NOW()"},
			}),
		}).Create(stats).Error
		if err != nil {
			return err
		}
	}

	// Upsert substitution patterns
	for _, sub := range subsMap {
		err := s.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "expected_phoneme"}, {Name: "actual_phoneme"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"occurrence_count": clause.Expr{SQL: "phoneme_substitutions.occurrence_count + ?", Vars: []interface{}{sub.OccurrenceCount}},
				"updated_at":       clause.Expr{SQL: "NOW()"},
			}),
		}).Create(sub).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// UserPhonemeStatsResponse contains aggregated phoneme stats for a user
type UserPhonemeStatsResponse struct {
	TotalPhonemes       int                   `json:"totalPhonemes"`
	OverallAccuracy     float64               `json:"overallAccuracy"`
	PhonemeStats        []PhonemeAccuracy     `json:"phonemeStats"`
	CommonSubstitutions []SubstitutionPattern `json:"commonSubstitutions"`
}

// PhonemeAccuracy represents a single phoneme's accuracy
type PhonemeAccuracy struct {
	Phoneme       string  `json:"phoneme"`
	TotalAttempts int     `json:"totalAttempts"`
	CorrectCount  int     `json:"correctCount"`
	DeletionCount int     `json:"deletionCount"`
	Accuracy      float64 `json:"accuracy"`
}

// SubstitutionPattern represents a common substitution error
type SubstitutionPattern struct {
	ExpectedPhoneme string `json:"expectedPhoneme"`
	ActualPhoneme   string `json:"actualPhoneme"`
	Count           int    `json:"count"`
}

// GetUserStats retrieves aggregated phoneme statistics for a user
func (s *PhonemeStatsService) GetUserStats(userID uuid.UUID) (*UserPhonemeStatsResponse, error) {
	var stats []models.PhonemeStats
	if err := s.DB.Where("user_id = ?", userID).Find(&stats).Error; err != nil {
		return nil, err
	}

	// Calculate totals
	var totalAttempts, totalCorrect int
	for _, stat := range stats {
		totalAttempts += stat.TotalAttempts
		totalCorrect += stat.CorrectCount
	}

	overallAccuracy := 0.0
	if totalAttempts > 0 {
		overallAccuracy = float64(totalCorrect) / float64(totalAttempts) * 100
	}

	// Get all phoneme stats for this user
	var phonemeStats []PhonemeAccuracy
	if err := s.DB.Model(&models.PhonemeStats{}).
		Select("phoneme, total_attempts, correct_count, deletion_count, (CAST(correct_count AS FLOAT) / CAST(total_attempts AS FLOAT) * 100) as accuracy").
		Where("user_id = ?", userID).
		Order("accuracy ASC").
		Scan(&phonemeStats).Error; err != nil {
		return nil, err
	}

	// Get common substitutions
	var substitutions []models.PhonemeSubstitution
	if err := s.DB.Where("user_id = ?", userID).
		Order("occurrence_count DESC").
		Limit(10).
		Find(&substitutions).Error; err != nil {
		return nil, err
	}

	commonSubs := make([]SubstitutionPattern, len(substitutions))
	for i, sub := range substitutions {
		commonSubs[i] = SubstitutionPattern{
			ExpectedPhoneme: sub.ExpectedPhoneme,
			ActualPhoneme:   sub.ActualPhoneme,
			Count:           sub.OccurrenceCount,
		}
	}

	return &UserPhonemeStatsResponse{
		TotalPhonemes:       totalAttempts,
		OverallAccuracy:     overallAccuracy,
		PhonemeStats:        phonemeStats,
		CommonSubstitutions: commonSubs,
	}, nil
}
