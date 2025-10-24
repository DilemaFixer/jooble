package differ

import "jooble-parser/internal/domain"

type Differ interface {
	Check(parsed []domain.Job) ([]domain.Job, error)
}
