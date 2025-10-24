package domain

import "strconv"

type Job struct {
	ID          int64    `json:"id"`
	ExternalID  string   `json:"external_id"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	City        string   `json:"city"`
	Salary      string   `json:"salary"`
	Link        string   `json:"link"`
	Description string   `json:"description"`
	WorkType    string   `json:"work_type"`
	Date        string   `json:"date"`
	Tags        []string `json:"tags"`
}

func (job Job) GetId() (int64, error) {
	return strconv.ParseInt(job.ExternalID, 10, 64)
}
