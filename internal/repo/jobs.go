package repo

import (
	"database/sql"
	"fmt"
	"jooble-parser/internal/domain"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type JobsRepository interface {
	GetJobs() ([]domain.Job, error)
	GetById(id int64) (*domain.Job, error)
	AddJob(job domain.Job) error
	UpdateJob(job domain.Job) error
	DeleteJob(id int64) error

	GetByExternalID(externalID string) (*domain.Job, error)
	JobExists(externalID string) (bool, error)
	Count() (int64, error)

	InitSchema() error
}

type SQLiteJobsRepository struct {
	db    *sql.DB
	count uint
}

func NewSQLiteJobsRepository(db *sql.DB) *SQLiteJobsRepository {
	return &SQLiteJobsRepository{db: db}
}

func (r *SQLiteJobsRepository) InitSchema() error {
	query := `
    CREATE TABLE IF NOT EXISTS jobs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        external_id TEXT UNIQUE,
        title TEXT NOT NULL,
        company TEXT,
        city TEXT,
        salary TEXT,
        link TEXT,
        description TEXT,
        work_type TEXT,
        date TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS job_tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        job_id INTEGER,
        tag TEXT,
        FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
    );
    
    CREATE INDEX IF NOT EXISTS idx_jobs_external_id ON jobs(external_id);
    CREATE INDEX IF NOT EXISTS idx_jobs_date ON jobs(date);
    CREATE INDEX IF NOT EXISTS idx_job_tags_job_id ON job_tags(job_id);
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteJobsRepository) GetJobs() ([]domain.Job, error) {
	query := `
    SELECT id, external_id, title, company, city, salary, link, description, work_type, date
    FROM jobs
    ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
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

		// Получаем теги
		tags, err := r.getJobTags(job.ID)
		if err != nil {
			return nil, err
		}
		job.Tags = tags

		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return jobs, nil
}

func (r *SQLiteJobsRepository) GetById(id int64) (*domain.Job, error) {
	query := `
    SELECT id, external_id, title, company, city, salary, link, description, work_type, date
    FROM jobs
    WHERE id = ?
    `

	var job domain.Job
	var externalID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
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

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if externalID.Valid {
		job.ExternalID = externalID.String
	}

	// Получаем теги
	tags, err := r.getJobTags(id)
	if err != nil {
		return nil, err
	}
	job.Tags = tags

	return &job, nil
}

func (r *SQLiteJobsRepository) AddJob(job domain.Job) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
    INSERT INTO jobs (external_id, title, company, city, salary, link, description, work_type, date)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	result, err := tx.Exec(query,
		job.ExternalID,
		job.Title,
		job.Company,
		job.City,
		job.Salary,
		job.Link,
		job.Description,
		job.WorkType,
		job.Date,
	)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Вставляем теги
	if err := r.insertJobTags(tx, jobID, job.Tags); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *SQLiteJobsRepository) UpdateJob(job domain.Job) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
    UPDATE jobs
    SET external_id = ?, title = ?, company = ?, city = ?, salary = ?,
        link = ?, description = ?, work_type = ?, date = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `

	result, err := tx.Exec(query,
		job.ExternalID,
		job.Title,
		job.Company,
		job.City,
		job.Salary,
		job.Link,
		job.Description,
		job.WorkType,
		job.Date,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job with id %d not found", job.ID)
	}

	// Удаляем старые теги
	_, err = tx.Exec("DELETE FROM job_tags WHERE job_id = ?", job.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old tags: %w", err)
	}

	// Вставляем новые теги
	if err := r.insertJobTags(tx, job.ID, job.Tags); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *SQLiteJobsRepository) DeleteJob(id int64) error {
	query := `DELETE FROM jobs WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job with id %d not found", id)
	}

	return nil
}

func (r *SQLiteJobsRepository) getJobTags(jobID int64) ([]string, error) {
	query := `SELECT tag FROM job_tags WHERE job_id = ?`

	rows, err := r.db.Query(query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tags, nil
}

func (r *SQLiteJobsRepository) insertJobTags(tx *sql.Tx, jobID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(tags))
	valueArgs := make([]interface{}, 0, len(tags)*2)

	for _, tag := range tags {
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, jobID, tag)
	}

	query := fmt.Sprintf("INSERT INTO job_tags (job_id, tag) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err := tx.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert tags: %w", err)
	}

	return nil
}

func (r *SQLiteJobsRepository) GetByExternalID(externalID string) (*domain.Job, error) {
	query := `
    SELECT id, external_id, title, company, city, salary, link, description, work_type, date
    FROM jobs
    WHERE external_id = ?
    `

	var job domain.Job
	var extID sql.NullString

	err := r.db.QueryRow(query, externalID).Scan(
		&job.ID,
		&extID,
		&job.Title,
		&job.Company,
		&job.City,
		&job.Salary,
		&job.Link,
		&job.Description,
		&job.WorkType,
		&job.Date,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job by external id: %w", err)
	}

	if extID.Valid {
		job.ExternalID = extID.String
	}

	tags, err := r.getJobTags(job.ID)
	if err != nil {
		return nil, err
	}
	job.Tags = tags

	return &job, nil
}

func (r *SQLiteJobsRepository) JobExists(externalID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM jobs WHERE external_id = ?)`
	err := r.db.QueryRow(query, externalID).Scan(&exists)
	return exists, err
}

func (r *SQLiteJobsRepository) Count() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM jobs`
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
