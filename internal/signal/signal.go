package signal

import "jooble-parser/internal/domain"

type UpdateSignal interface {
	Signal(job []domain.Job) error
}
