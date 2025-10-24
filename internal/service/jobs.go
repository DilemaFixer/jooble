package service

import (
	"database/sql"
	"fmt"
	"jooble-parser/internal/domain"
	"jooble-parser/internal/repo"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type JobService interface {
	GetJobs() ([]domain.Job, error)
	GetById(id int64) (*domain.Job, error)
	AddJob(job domain.Job) error
	UpdateJob(job domain.Job) error
	DeleteJob(id int64) error

	GetByExternalID(externalID string) (*domain.Job, error)
	JobExists(externalID string) (bool, error)
	Count() (int64, error)

	AddJobWithLimit(job domain.Job) error
	CleanupOldJobs() error
}

type SqliteJobService struct {
	repo         repo.JobsRepository
	db           *sql.DB
	maxCount     int64
	clearingStep int64
	logger       *zap.Logger
}

func NewSqliteRepoService(pathToDb string, maxCount uint, clearingStep uint, logger *zap.Logger) (JobService, error) {
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repository := repo.NewSQLiteJobsRepository(db)

	if err := repository.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	service := &SqliteJobService{
		repo:         repository,
		maxCount:     int64(maxCount),
		db:           db,
		logger:       logger,
		clearingStep: int64(clearingStep),
	}

	return service, nil
}

func (s *SqliteJobService) GetJobs() ([]domain.Job, error) {
	return s.repo.GetJobs()
}

func (s *SqliteJobService) GetById(id int64) (*domain.Job, error) {
	return s.repo.GetById(id)
}

func (s *SqliteJobService) AddJob(job domain.Job) error {
	return s.repo.AddJob(job)
}

func (s *SqliteJobService) AddJobWithLimit(job domain.Job) error {
	exists, err := s.repo.JobExists(job.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to check job existence: %w", err)
	}

	if exists {
		return nil
	}

	count, err := s.repo.Count()
	if err != nil {
		return fmt.Errorf("failed to get jobs count: %w", err)
	}

	if count >= s.maxCount {
		if err := s.deleteOldestJobs(count - s.maxCount + 1); err != nil {
			return fmt.Errorf("failed to delete old jobs: %w", err)
		}
	}

	return s.repo.AddJob(job)
}

func (s *SqliteJobService) UpdateJob(job domain.Job) error {
	return s.repo.UpdateJob(job)
}

func (s *SqliteJobService) DeleteJob(id int64) error {
	return s.repo.DeleteJob(id)
}

func (s *SqliteJobService) GetByExternalID(externalID string) (*domain.Job, error) {
	return s.repo.GetByExternalID(externalID)
}

func (s *SqliteJobService) JobExists(externalID string) (bool, error) {
	return s.repo.JobExists(externalID)
}

func (s *SqliteJobService) Count() (int64, error) {
	return s.repo.Count()
}

func (s *SqliteJobService) CleanupOldJobs() error {
	count, err := s.repo.Count()
	if err != nil {
		return fmt.Errorf("failed to get jobs count: %w", err)
	}

	if count <= s.maxCount {
		return nil
	}

	toDelete := count - s.clearingStep
	return s.deleteOldestJobs(toDelete)
}

func (s *SqliteJobService) deleteOldestJobs(count int64) error {
	query := `
	DELETE FROM jobs 
	WHERE id IN (
		SELECT id FROM jobs 
		ORDER BY created_at ASC 
		LIMIT ?
	)
	`

	result, err := s.db.Exec(query, count)
	if err != nil {
		return fmt.Errorf("failed to delete oldest jobs: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if deleted > 0 {
		fmt.Printf("Deleted %d old jobs\n", deleted)
	}

	return nil
}

func (s *SqliteJobService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SqliteJobService) GetOldestJobs(limit int) ([]domain.Job, error) {
	query := `
	SELECT id, external_id, title, company, city, salary, link, description, work_type, date, created_at
	FROM jobs
	ORDER BY created_at ASC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query oldest jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var externalID sql.NullString
		var createdAt string

		err := rows.Scan(
			&job.ID,
			&externalID,
			&job.Title,
			&job.Company,
			&job.City,
			&job.Salary,
			&job.Link,
			&job.Description,
			&job.WorkType,
			&job.Date,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if externalID.Valid {
			job.ExternalID = externalID.String
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *SqliteJobService) GetNewestJobs(limit int) ([]domain.Job, error) {
	query := `
	SELECT id, external_id, title, company, city, salary, link, description, work_type, date
	FROM jobs
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query newest jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var externalID sql.NullString

		err := rows.Scan(
			&job.ID,
			&externalID,
			&job.Title,
			&job.Company,
			&job.City,
			&job.Salary,
			&job.Link,
			&job.Description,
			&job.WorkType,
			&job.Date,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if externalID.Valid {
			job.ExternalID = externalID.String
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
