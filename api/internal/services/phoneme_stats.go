package services

import (
	"ling-app/api/internal/client"
	"ling-app/api/internal/db"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"

	"github.com/google/uuid"
)

// PhonemeStatsService handles phoneme statistics aggregation
type PhonemeStatsService struct {
	db       *db.DB
	exec     repository.Executor
	statsRepo repository.PhonemeStatsRepository
	subsRepo  repository.PhonemeSubstitutionRepository
}

// NewPhonemeStatsService creates a new phoneme stats service
func NewPhonemeStatsService(
	database *db.DB,
	statsRepo repository.PhonemeStatsRepository,
	subsRepo repository.PhonemeSubstitutionRepository,
) *PhonemeStatsService {
	return &PhonemeStatsService{
		db:        database,
		exec:      database.DB,
		statsRepo: statsRepo,
		subsRepo:  subsRepo,
	}
}

// RecordPhonemeResults processes phoneme details from pronunciation analysis
// and updates the user's aggregate statistics
func (s *PhonemeStatsService) RecordPhonemeResults(userID uuid.UUID, phonemeDetails []client.PhonemeDetail) error {
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

	// Upsert phoneme stats using repository
	for _, stats := range statsMap {
		if err := s.statsRepo.Upsert(s.exec, stats); err != nil {
			return err
		}
	}

	// Upsert substitution patterns using repository
	for _, sub := range subsMap {
		if err := s.subsRepo.Upsert(s.exec, sub); err != nil {
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

// PhonemeAccuracy represents a single phoneme's accuracy (for API responses)
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
	// Get all stats for this user
	stats, err := s.statsRepo.FindByUserID(s.exec, userID)
	if err != nil {
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

	// Get accuracy ranking from repository
	repoAccuracy, err := s.statsRepo.GetAccuracyRanking(s.exec, userID)
	if err != nil {
		return nil, err
	}

	// Convert repository PhonemeAccuracy to service PhonemeAccuracy
	phonemeStats := make([]PhonemeAccuracy, len(repoAccuracy))
	for i, pa := range repoAccuracy {
		phonemeStats[i] = PhonemeAccuracy{
			Phoneme:       pa.Phoneme,
			TotalAttempts: pa.TotalAttempts,
			CorrectCount:  pa.CorrectCount,
			DeletionCount: pa.DeletionCount,
			Accuracy:      pa.Accuracy,
		}
	}

	// Get common substitutions
	substitutions, err := s.subsRepo.FindTopByUserID(s.exec, userID, 10)
	if err != nil {
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
