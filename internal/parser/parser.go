package parser

import (
	"fmt"
	"jooble-parser/internal/domain"
	"jooble-parser/internal/parser/setters"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type JobParser struct {
	propsSetter []setters.PropSeter
	logger      *zap.Logger
}

func NewJobParser(logger *zap.Logger, props ...setters.PropSeter) *JobParser {
	if len(props) == 0 {
		panic("No job props setters set")
	}

	return &JobParser{
		propsSetter: props,
		logger:      logger,
	}
}

func (p *JobParser) Parse(html string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var jobs []domain.Job
	var parseErrors []error

	doc.Find("ul.kiBEcn").
		Find(`div[data-test-name="_jobCard"]`).
		Each(func(i int, s *goquery.Selection) {
			job := &domain.Job{}

			for _, propSeter := range p.propsSetter {
				if err := propSeter(job, s); err != nil {
					parseErrors = append(parseErrors, fmt.Errorf("job %d: %w", i, err))
					return
				}
			}

			jobs = append(jobs, *job)
		})

	if len(parseErrors) > 0 {
		return jobs, fmt.Errorf("parsing errors: %v", parseErrors)
	}

	return jobs, nil
}
