package repositories

import (
	"app/internal/domain"
	"encoding/json"
	"gorm.io/gorm"
)

// EventRepository interface
//
//go:generate mockery --name=EventRepository --output=../mocks --structname=EventRepositoryMock
type EventRepository interface {
	Save(eventType string, payload interface{}) error
	GetByID(id int64) (*domain.Event, error)
	ListUnprocessed(limit int) ([]domain.Event, error)
	MarkProcessedBatch(ids []int64) error
}

// EventRepositoryImpl implementation
type EventRepositoryImpl struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &EventRepositoryImpl{db: db}
}

func (r *EventRepositoryImpl) Save(eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := &domain.Event{
		Type:      eventType,
		Payload:   string(data),
		Processed: false,
	}

	return r.db.Create(event).Error
}

func (r *EventRepositoryImpl) GetByID(id int64) (*domain.Event, error) {
	var event domain.Event
	if err := r.db.First(&event, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepositoryImpl) ListUnprocessed(limit int) ([]domain.Event, error) {
	var events []domain.Event
	query := r.db.Where("processed = ?", false).Order("id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepositoryImpl) MarkProcessedBatch(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&domain.Event{}).
		Where("id IN ?", ids).
		Update("processed", true).Error
}
