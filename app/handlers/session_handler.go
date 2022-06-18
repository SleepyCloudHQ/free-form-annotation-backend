package handlers

import (
	"backend/app/models"
	"gorm.io/gorm"
	"time"
)

type SessionStats struct {
	TotalSamples     int64 `json:"total_samples"`
	CompletedSamples int64 `json:"completed_samples"`
	PendingSamples   int64 `json:"pending_samples"`
}

type SessionData struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Type      models.SessionType `json:"type"`
	CreatedAt time.Time          `json:"created_at"`
	Stats     *SessionStats      `json:"stats"`
}

type SessionHandler struct {
	DB *gorm.DB
}

func NewSessionHandler(db *gorm.DB) *SessionHandler {
	return &SessionHandler{
		DB: db,
	}
}

func (s *SessionHandler) GetSessions() *[]SessionData {
	var sessions []models.Session
	s.DB.Find(&sessions)

	result := make([]SessionData, len(sessions))
	for i, session := range sessions {
		result[i] = *s.mapSessionToSessionData(&session)
	}
	return &result
}

func (s *SessionHandler) GetSession(id uint) (*SessionData, error) {
	session := &models.Session{}
	if dbErr := s.DB.First(session, id).Error; dbErr != nil {
		return nil, dbErr
	}

	return s.mapSessionToSessionData(session), nil
}

func (s *SessionHandler) mapSessionToSessionData(session *models.Session) *SessionData {
	return &SessionData{
		ID:        session.ID,
		Name:      session.Name,
		Type:      session.Type,
		CreatedAt: session.CreatedAt,
		Stats:     s.getSessionsStats(session),
	}
}

func (s *SessionHandler) getSessionsStats(session *models.Session) *SessionStats {
	stats := &SessionStats{
		TotalSamples:     0,
		CompletedSamples: 0,
		PendingSamples:   0,
	}

	completed := []models.StatusType{models.Accepted, models.Rejected, models.Uncertain}

	stats.TotalSamples = s.DB.Model(session).Association("Samples").Count()
	stats.CompletedSamples = s.DB.Model(session).Where("status IN ?", completed).Association("Samples").Count()
	stats.PendingSamples = stats.TotalSamples - stats.CompletedSamples

	return stats
}
