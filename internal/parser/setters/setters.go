package setters

import (
	"jooble-parser/internal/domain"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var AllSetters = []PropSeter{
	IdSeter,
	LinkSeter,
	CompanySeter,
	CitySeter,
	SalarySeter,
	WorkTypeSeter,
	DateSeter,
	DescriptionSeter,
	TagsSeter,
}

type PropSeter func(job *domain.Job, selection *goquery.Selection) error

func IdSeter(job *domain.Job, selection *goquery.Selection) error {
	if id, exists := selection.Attr("id"); exists {
		job.ExternalID = id
	}
	return nil
}

func LinkSeter(job *domain.Job, selection *goquery.Selection) error {
	linkSel := selection.Find("h2 a")
	job.Title = strings.TrimSpace(linkSel.Text())
	if href, ok := linkSel.Attr("href"); ok {
		job.Link = strings.TrimSpace(href)
	}

	return nil
}

func DescriptionSeter(job *domain.Job, selection *goquery.Selection) error {
	job.Description = strings.TrimSpace(selection.Find("div.GEyos4.e9eiOZ").Text())
	return nil
}

func CompanySeter(job *domain.Job, selection *goquery.Selection) error {
	job.Company = strings.TrimSpace(selection.Find(`p[data-test-name="_companyName"]`).Text())
	return nil
}

func CitySeter(job *domain.Job, selection *goquery.Selection) error {
	job.City = strings.TrimSpace(selection.Find("div.caption.NTRJBV").Text())
	return nil
}

func SalarySeter(job *domain.Job, selection *goquery.Selection) error {
	job.Salary = strings.TrimSpace(selection.Find("p.b97WnG").Text())
	return nil
}

func WorkTypeSeter(job *domain.Job, selection *goquery.Selection) error {
	job.WorkType = strings.TrimSpace(selection.Find("p._1dYE+p").Text())
	return nil
}

func DateSeter(job *domain.Job, selection *goquery.Selection) error {
	job.Date = strings.TrimSpace(selection.Find("div.GEyos4.e9eiOZ span:first-child").Text())
	return nil
}

func TagsSeter(job *domain.Job, selection *goquery.Selection) error {
	selection.Find("div.K8ZLnh.tag").Each(func(_ int, tagSel *goquery.Selection) {
		tag := strings.TrimSpace(tagSel.Text())
		if tag != "" {
			job.Tags = append(job.Tags, tag)
		}
	})
	return nil
}
