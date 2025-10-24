package signal

import (
	"fmt"
	"jooble-parser/internal/domain"
)

type LogUpdateSignal struct {
	n int
}

func NewLogUpdateSignal() UpdateSignal {
	return &LogUpdateSignal{}
}

func (signal *LogUpdateSignal) Signal(job []domain.Job) error {
	fmt.Println("Log Update Signal")
	fmt.Println(job)
	return nil
}
