package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ling-app/api/internal/models"
)

// Executor is an interface that both *gorm.DB and *gorm.Tx satisfy.
// This allows repository methods to work with either a database connection
// or a transaction, enabling atomic operations across multiple repositories.
type Executor interface {
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(query string, args ...interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Limit(limit int) *gorm.DB
	Model(value interface{}) *gorm.DB
	Update(column string, value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Scan(dest interface{}) *gorm.DB
	Clauses(conds ...clause.Expression) *gorm.DB
}

// UserRepository handles user persistence.
type UserRepository interface {
	FindByEmail(exec Executor, email string) (*models.User, error)
	FindByGoogleID(exec Executor, googleID string) (*models.User, error)
	FindByGitHubID(exec Executor, githubID string) (*models.User, error)
	Create(exec Executor, user *models.User) error
	Save(exec Executor, user *models.User) error
}

// SessionRepository handles session persistence.
type SessionRepository interface {
	Create(exec Executor, session *models.Session) error
	FindByIDWithUser(exec Executor, token string) (*models.Session, error)
	DeleteByID(exec Executor, token string) error
	DeleteByUserID(exec Executor, userID uuid.UUID) error
	DeleteExpiredBefore(exec Executor, t time.Time) (int64, error)
}

// CreditsRepository handles credits persistence.
type CreditsRepository interface {
	FindByUserID(exec Executor, userID uuid.UUID) (*models.Credits, error)
	Create(exec Executor, credits *models.Credits) error
	Save(exec Executor, credits *models.Credits) error
	UpdateAllowance(exec Executor, userID uuid.UUID, allowance int) error
}

// CreditTransactionRepository handles credit transaction persistence.
type CreditTransactionRepository interface {
	Create(exec Executor, tx *models.CreditTransaction) error
	FindByUserID(exec Executor, userID uuid.UUID, limit int) ([]models.CreditTransaction, error)
}

// PhonemeStatsRepository handles phoneme statistics persistence.
type PhonemeStatsRepository interface {
	Upsert(exec Executor, stats *models.PhonemeStats) error
	FindByUserID(exec Executor, userID uuid.UUID) ([]models.PhonemeStats, error)
	GetAccuracyRanking(exec Executor, userID uuid.UUID) ([]PhonemeAccuracy, error)
}

// PhonemeAccuracy represents a single phoneme's accuracy stats.
type PhonemeAccuracy struct {
	Phoneme       string
	TotalAttempts int
	CorrectCount  int
	DeletionCount int
	Accuracy      float64
}

// PhonemeSubstitutionRepository handles phoneme substitution patterns persistence.
type PhonemeSubstitutionRepository interface {
	Upsert(exec Executor, sub *models.PhonemeSubstitution) error
	FindTopByUserID(exec Executor, userID uuid.UUID, limit int) ([]models.PhonemeSubstitution, error)
}

// SubscriptionRepository handles subscription persistence.
type SubscriptionRepository interface {
	FindByUserID(exec Executor, userID uuid.UUID) (*models.Subscription, error)
	FindByStripeCustomerID(exec Executor, customerID string) (*models.Subscription, error)
	FindByStripeSubscriptionID(exec Executor, subscriptionID string) (*models.Subscription, error)
	Create(exec Executor, sub *models.Subscription) error
	Save(exec Executor, sub *models.Subscription) error
	UpdateStatus(exec Executor, subscriptionID string, status string) error
}

// ThreadRepository handles thread persistence.
type ThreadRepository interface {
	Create(exec Executor, thread *models.Thread) error
	FindByID(exec Executor, id uuid.UUID) (*models.Thread, error)
	FindByIDWithMessages(exec Executor, id uuid.UUID) (*models.Thread, error)
	FindByUserID(exec Executor, userID uuid.UUID) ([]models.Thread, error)
	FindArchivedByUserID(exec Executor, userID uuid.UUID) ([]models.Thread, error)
	FindByIDAndUserID(exec Executor, id, userID uuid.UUID) (*models.Thread, error)
	FindByIDAndUserIDWithMessages(exec Executor, id, userID uuid.UUID) (*models.Thread, error)
	Save(exec Executor, thread *models.Thread) error
	Delete(exec Executor, thread *models.Thread) error
	UpdateName(exec Executor, id uuid.UUID, name string) error
}

// MessageRepository handles message persistence.
type MessageRepository interface {
	Create(exec Executor, message *models.Message) error
	FindByID(exec Executor, id uuid.UUID) (*models.Message, error)
	FindByThreadID(exec Executor, threadID uuid.UUID) ([]models.Message, error)
	UpdatePronunciationStatus(exec Executor, id uuid.UUID, status string) error
	UpdatePronunciationAnalysis(exec Executor, id uuid.UUID, status string, analysis models.JSONMap, updatedAt time.Time) error
	UpdatePronunciationError(exec Executor, id uuid.UUID, status string, errMsg string, updatedAt time.Time) error
}
