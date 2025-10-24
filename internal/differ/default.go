package differ

import (
	"jooble-parser/internal/domain"
	"jooble-parser/internal/service"
)

type SqliteDiffer struct {
	repository service.JobService
}

func NewDefaultDiffer(service service.JobService) Differ {
	return &SqliteDiffer{
		repository: service,
	}
}

func (d *SqliteDiffer) Check(parsed []domain.Job) ([]domain.Job, error) {
	count, err := d.repository.Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		if err := d.markAsExisting(parsed); err != nil {
			return nil, err
		}
		return parsed, nil
	}

	previous, err := d.repository.GetJobs()
	if err != nil {
		return nil, err
	}

	result := []domain.Job{}
	for _, job := range parsed {
		if !checkExisting(job, previous) {
			result = append(result, job)
		}
	}

	if err := d.markAsExisting(result); err != nil {
		return nil, err
	}

	return result, nil
}

func checkExisting(job domain.Job, existing []domain.Job) bool {
	for _, existingJob := range existing {
		if job.ExternalID == existingJob.ExternalID {
			return true
		}
	}
	return false
}

func (d *SqliteDiffer) markAsExisting(jobs []domain.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	for _, job := range jobs {
		if err := d.repository.AddJobWithLimit(job); err != nil {
			return err
		}
	}
	return nil
}
