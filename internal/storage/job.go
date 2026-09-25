package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"video-downloader-bot/internal/models"
)

type JobStore struct {
	db *gorm.DB
}

func NewJobStorage(db *gorm.DB) *JobStore {
	return &JobStore{db: db}
}

func (s *JobStore) Create(ctx context.Context, job *models.Job) error {
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("failed to save job for %q: %w", job.Link, err)
	}

	return nil
}

func (s *JobStore) Get(ctx context.Context, id uint) (*models.Job, error) {
	var job models.Job
	if err := s.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, fmt.Errorf("job %d not found: %w", id, err)
	}

	return &job, nil
}

// Delete removes the job and reports whether it was still there, so two concurrent retry clicks can't both start a download
func (s *JobStore) Delete(ctx context.Context, id uint) (bool, error) {
	res := s.db.WithContext(ctx).Delete(&models.Job{}, id)
	if res.Error != nil {
		return false, fmt.Errorf("failed to delete job %d: %w", id, res.Error)
	}

	return res.RowsAffected > 0, nil
}

func (s *JobStore) MarkStale(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&models.Job{}).Where("id = ?", id).Update("stale_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to mark job %d stale: %w", id, err)
	}

	return nil
}

// ListStuck returns jobs started before the given time that nobody has finished or marked stale
func (s *JobStore) ListStuck(ctx context.Context, before time.Time) ([]models.Job, error) {
	var jobs []models.Job
	if err := s.db.WithContext(ctx).Where("stale_at IS NULL AND created_at < ?", before).Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to list stuck jobs: %w", err)
	}

	return jobs, nil
}

func (s *JobStore) DeleteOlderThan(ctx context.Context, before time.Time) error {
	if err := s.db.WithContext(ctx).Where("created_at < ?", before).Delete(&models.Job{}).Error; err != nil {
		return fmt.Errorf("failed to delete old jobs: %w", err)
	}

	return nil
}
