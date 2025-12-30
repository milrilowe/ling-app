package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
)

// subscriptionRepository implements SubscriptionRepository using GORM.
type subscriptionRepository struct{}

// NewSubscriptionRepository creates a new GORM-backed subscription repository.
func NewSubscriptionRepository() SubscriptionRepository {
	return &subscriptionRepository{}
}

func (r *subscriptionRepository) FindByUserID(exec Executor, userID uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := exec.Where("user_id = ?", userID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) FindByStripeCustomerID(exec Executor, customerID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := exec.Where("stripe_customer_id = ?", customerID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) FindByStripeSubscriptionID(exec Executor, subscriptionID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := exec.Where("stripe_subscription_id = ?", subscriptionID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Create(exec Executor, sub *models.Subscription) error {
	return exec.Create(sub).Error
}

func (r *subscriptionRepository) Save(exec Executor, sub *models.Subscription) error {
	return exec.Save(sub).Error
}

func (r *subscriptionRepository) UpdateStatus(exec Executor, subscriptionID string, status string) error {
	return exec.Model(&models.Subscription{}).
		Where("stripe_subscription_id = ?", subscriptionID).
		Update("status", status).Error
}
