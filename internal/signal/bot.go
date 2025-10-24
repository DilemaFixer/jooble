package signal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jooble-parser/internal/config"
	"jooble-parser/internal/domain"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type BotUpdateSignal struct {
	token      string
	customerId int64
	logger     *zap.Logger
}

func NewBotUpdateSignal(cfg *config.Config, logger *zap.Logger) *BotUpdateSignal {
	return &BotUpdateSignal{
		token:      cfg.Signal.Token,
		customerId: cfg.Signal.CustomerId,
		logger:     logger,
	}
}

func (u *BotUpdateSignal) Signal(jobs []domain.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	for _, job := range jobs {
		message := u.formatJobMessage(job)
		if err := u.sendMessage(message, job.Link); err != nil {
			return fmt.Errorf("failed to send job %d: %w", job.ID, err)
		}
	}
	u.logger.Debug("Successfully sent update signal", zap.Int("jobs", len(jobs)))
	return nil
}

func (u *BotUpdateSignal) formatJobMessage(job domain.Job) string {
	var sb strings.Builder

	sb.WriteString("🔔 <b>Новая вакансия</b>\n\n")
	sb.WriteString(fmt.Sprintf("📋 <b>%s</b>\n", escapeHTML(job.Title)))
	sb.WriteString(fmt.Sprintf("🏢 %s\n", escapeHTML(job.Company)))

	if job.City != "" {
		sb.WriteString(fmt.Sprintf("📍 %s\n", escapeHTML(job.City)))
	}

	if job.Salary != "" {
		sb.WriteString(fmt.Sprintf("💰 %s\n", escapeHTML(job.Salary)))
	}

	if job.WorkType != "" {
		sb.WriteString(fmt.Sprintf("💼 %s\n", escapeHTML(job.WorkType)))
	}

	if job.Date != "" {
		sb.WriteString(fmt.Sprintf("📅 %s\n", escapeHTML(job.Date)))
	}

	if job.Description != "" {
		description := job.Description
		if len(description) > 400 {
			description = description[:397] + "..."
		}
		sb.WriteString(fmt.Sprintf("\n%s\n", escapeHTML(description)))
	}

	if len(job.Tags) > 0 {
		sb.WriteString("\n")
		hashtags := make([]string, 0, len(job.Tags))
		for _, tag := range job.Tags {
			// Убираем пробелы и спецсимволы из хештегов
			cleanTag := strings.ReplaceAll(tag, " ", "_")
			cleanTag = strings.ReplaceAll(cleanTag, ".", "")
			cleanTag = strings.ReplaceAll(cleanTag, ",", "")
			hashtags = append(hashtags, "#"+cleanTag)
		}
		sb.WriteString(strings.Join(hashtags, " "))
	}

	return sb.String()
}

func (u *BotUpdateSignal) sendMessage(text string, link string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", u.token)

	payload := map[string]interface{}{
		"chat_id":    u.customerId,
		"text":       text,
		"parse_mode": "HTML",
	}

	// Добавляем кнопку-ссылку, если есть link
	if link != "" {
		payload["reply_markup"] = map[string]interface{}{
			"inline_keyboard": [][]map[string]string{
				{
					{
						"text": "🔗 Посмотреть вакансию",
						"url":  link,
					},
				},
			},
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		return fmt.Errorf("telegram api error: %v", result)
	}

	return nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
