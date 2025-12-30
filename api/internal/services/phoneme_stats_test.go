package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/client"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/repository/mocks"
)

// NewPhonemeStatsServiceForTest creates a PhonemeStatsService with injected dependencies for testing.
func NewPhonemeStatsServiceForTest(
	exec repository.Executor,
	statsRepo repository.PhonemeStatsRepository,
	subsRepo repository.PhonemeSubstitutionRepository,
) *PhonemeStatsService {
	return &PhonemeStatsService{
		db:        nil,
		exec:      exec,
		statsRepo: statsRepo,
		subsRepo:  subsRepo,
	}
}

func TestPhonemeStatsService_RecordPhonemeResults(t *testing.T) {
	userID := uuid.New()

	t.Run("returns nil for empty phoneme details", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, []client.PhonemeDetail{})

		assert.NoError(t, err)
		// No repository calls should be made
		statsRepo.AssertNotCalled(t, "Upsert")
		subsRepo.AssertNotCalled(t, "Upsert")
	})

	t.Run("records match phonemes correctly", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "æ", Actual: "æ", Type: "match"},
			{Expected: "æ", Actual: "æ", Type: "match"},
			{Expected: "ɪ", Actual: "ɪ", Type: "match"},
		}

		// Expect upserts for both phonemes
		statsRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *models.PhonemeStats) bool {
			if s.Phoneme == "æ" {
				return s.UserID == userID && s.TotalAttempts == 2 && s.CorrectCount == 2 && s.DeletionCount == 0
			}
			if s.Phoneme == "ɪ" {
				return s.UserID == userID && s.TotalAttempts == 1 && s.CorrectCount == 1 && s.DeletionCount == 0
			}
			return false
		})).Return(nil).Times(2)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.NoError(t, err)
		statsRepo.AssertExpectations(t)
	})

	t.Run("records deletion phonemes correctly", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "θ", Actual: "", Type: "delete"},
		}

		statsRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *models.PhonemeStats) bool {
			return s.Phoneme == "θ" && s.TotalAttempts == 1 && s.CorrectCount == 0 && s.DeletionCount == 1
		})).Return(nil)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.NoError(t, err)
		statsRepo.AssertExpectations(t)
	})

	t.Run("records substitution phonemes and patterns", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "θ", Actual: "f", Type: "substitute"},
			{Expected: "θ", Actual: "f", Type: "substitute"},
			{Expected: "ð", Actual: "d", Type: "substitute"},
		}

		// Stats for both phonemes
		statsRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *models.PhonemeStats) bool {
			return s.Phoneme == "θ" || s.Phoneme == "ð"
		})).Return(nil).Times(2)

		// Substitution patterns
		subsRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *models.PhonemeSubstitution) bool {
			if s.ExpectedPhoneme == "θ" && s.ActualPhoneme == "f" {
				return s.OccurrenceCount == 2
			}
			if s.ExpectedPhoneme == "ð" && s.ActualPhoneme == "d" {
				return s.OccurrenceCount == 1
			}
			return false
		})).Return(nil).Times(2)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.NoError(t, err)
		statsRepo.AssertExpectations(t)
		subsRepo.AssertExpectations(t)
	})

	t.Run("skips insert type phonemes", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "", Actual: "ə", Type: "insert"},
			{Expected: "æ", Actual: "æ", Type: "match"},
		}

		// Only the match phoneme should be recorded
		statsRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *models.PhonemeStats) bool {
			return s.Phoneme == "æ" && s.TotalAttempts == 1 && s.CorrectCount == 1
		})).Return(nil)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.NoError(t, err)
		statsRepo.AssertExpectations(t)
	})

	t.Run("skips phonemes with empty expected", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "", Actual: "ə", Type: "match"},
		}

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.NoError(t, err)
		statsRepo.AssertNotCalled(t, "Upsert")
	})

	t.Run("returns error when stats upsert fails", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "æ", Actual: "æ", Type: "match"},
		}

		dbError := errors.New("database error")
		statsRepo.On("Upsert", mock.Anything, mock.Anything).Return(dbError)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
	})

	t.Run("returns error when substitution upsert fails", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeDetails := []client.PhonemeDetail{
			{Expected: "θ", Actual: "f", Type: "substitute"},
		}

		statsRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil)
		dbError := errors.New("substitution db error")
		subsRepo.On("Upsert", mock.Anything, mock.Anything).Return(dbError)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		err := service.RecordPhonemeResults(userID, phonemeDetails)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
	})
}

func TestPhonemeStatsService_GetUserStats(t *testing.T) {
	userID := uuid.New()

	t.Run("returns stats for user with data", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		phonemeStats := []models.PhonemeStats{
			{UserID: userID, Phoneme: "æ", TotalAttempts: 10, CorrectCount: 8, DeletionCount: 1},
			{UserID: userID, Phoneme: "ɪ", TotalAttempts: 5, CorrectCount: 5, DeletionCount: 0},
		}

		accuracyRanking := []repository.PhonemeAccuracy{
			{Phoneme: "æ", TotalAttempts: 10, CorrectCount: 8, DeletionCount: 1, Accuracy: 80.0},
			{Phoneme: "ɪ", TotalAttempts: 5, CorrectCount: 5, DeletionCount: 0, Accuracy: 100.0},
		}

		substitutions := []models.PhonemeSubstitution{
			{UserID: userID, ExpectedPhoneme: "θ", ActualPhoneme: "f", OccurrenceCount: 5},
		}

		statsRepo.On("FindByUserID", mock.Anything, userID).Return(phonemeStats, nil)
		statsRepo.On("GetAccuracyRanking", mock.Anything, userID).Return(accuracyRanking, nil)
		subsRepo.On("FindTopByUserID", mock.Anything, userID, 10).Return(substitutions, nil)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		result, err := service.GetUserStats(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 15, result.TotalPhonemes) // 10 + 5
		// Overall accuracy: 13 correct / 15 total * 100 = 86.67%
		assert.InDelta(t, 86.67, result.OverallAccuracy, 0.01)
		assert.Len(t, result.PhonemeStats, 2)
		assert.Len(t, result.CommonSubstitutions, 1)
		assert.Equal(t, "θ", result.CommonSubstitutions[0].ExpectedPhoneme)
		assert.Equal(t, "f", result.CommonSubstitutions[0].ActualPhoneme)
		assert.Equal(t, 5, result.CommonSubstitutions[0].Count)
	})

	t.Run("returns empty stats for user with no data", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		statsRepo.On("FindByUserID", mock.Anything, userID).Return([]models.PhonemeStats{}, nil)
		statsRepo.On("GetAccuracyRanking", mock.Anything, userID).Return([]repository.PhonemeAccuracy{}, nil)
		subsRepo.On("FindTopByUserID", mock.Anything, userID, 10).Return([]models.PhonemeSubstitution{}, nil)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		result, err := service.GetUserStats(userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.TotalPhonemes)
		assert.Equal(t, 0.0, result.OverallAccuracy)
		assert.Len(t, result.PhonemeStats, 0)
		assert.Len(t, result.CommonSubstitutions, 0)
	})

	t.Run("returns error when FindByUserID fails", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		dbError := errors.New("database error")
		statsRepo.On("FindByUserID", mock.Anything, userID).Return([]models.PhonemeStats{}, dbError)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		result, err := service.GetUserStats(userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
	})

	t.Run("returns error when GetAccuracyRanking fails", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		statsRepo.On("FindByUserID", mock.Anything, userID).Return([]models.PhonemeStats{}, nil)
		dbError := errors.New("ranking error")
		statsRepo.On("GetAccuracyRanking", mock.Anything, userID).Return([]repository.PhonemeAccuracy{}, dbError)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		result, err := service.GetUserStats(userID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when FindTopByUserID fails", func(t *testing.T) {
		statsRepo := new(mocks.MockPhonemeStatsRepository)
		subsRepo := new(mocks.MockPhonemeSubstitutionRepository)

		statsRepo.On("FindByUserID", mock.Anything, userID).Return([]models.PhonemeStats{}, nil)
		statsRepo.On("GetAccuracyRanking", mock.Anything, userID).Return([]repository.PhonemeAccuracy{}, nil)
		dbError := errors.New("substitution error")
		subsRepo.On("FindTopByUserID", mock.Anything, userID, 10).Return([]models.PhonemeSubstitution{}, dbError)

		service := NewPhonemeStatsServiceForTest(nil, statsRepo, subsRepo)
		result, err := service.GetUserStats(userID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
