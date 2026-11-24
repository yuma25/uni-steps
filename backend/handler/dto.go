package handler

import (
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// UserResponse はクライアントへ返却するユーザー情報のDTOである．
// トークン類（アクセストークン等）はセキュリティ保護のため除外している．
type UserResponse struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Email        string           `json:"email"`
	HasPushToken bool             `json:"has_push_token"`
	Groups       []*GroupResponse `json:"groups,omitempty"`
}

func ToUserResponse(u *domain.User) *UserResponse {
	if u == nil {
		return nil
	}
	resp := &UserResponse{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		HasPushToken: u.WebPushToken != "",
	}
	for _, g := range u.Groups {
		resp.Groups = append(resp.Groups, ToGroupResponse(g))
	}
	return resp
}

// GroupResponse はクライアントへ返却するグループ情報のDTOである．
type GroupResponse struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	OwnerID            string          `json:"owner_id"`
	LineChannelToken   string          `json:"line_channel_token,omitempty"`
	LineGroupID        string          `json:"line_group_id,omitempty"`
	LastSyncedAt       time.Time       `json:"last_synced_at"`
	LMSLastUpdatedAt   time.Time       `json:"lms_last_updated_at"`
	InviteCode         string          `json:"invite_code"`
	RemindIntervals    []int           `json:"remind_intervals"`
	AICharacter        string          `json:"ai_character"`
	SummaryMorningTime string          `json:"summary_morning_time"`
	SummaryEveningTime string          `json:"summary_evening_time"`
	Users              []*UserResponse `json:"users,omitempty"`
}

func ToGroupResponse(g *domain.Group) *GroupResponse {
	if g == nil {
		return nil
	}
	resp := &GroupResponse{
		ID:                 g.ID,
		Name:               g.Name,
		OwnerID:            g.OwnerID,
		LineChannelToken:   g.LineChannelToken,
		LineGroupID:        g.LineGroupID,
		LastSyncedAt:       g.LastSyncedAt,
		LMSLastUpdatedAt:   g.LMSLastUpdatedAt,
		InviteCode:         g.InviteCode,
		RemindIntervals:    g.RemindIntervals,
		AICharacter:        g.AICharacter,
		SummaryMorningTime: g.SummaryMorningTime,
		SummaryEveningTime: g.SummaryEveningTime,
	}
	for _, u := range g.Users {
		resp.Users = append(resp.Users, &UserResponse{
			ID:           u.ID,
			Name:         u.Name,
			Email:        u.Email,
			HasPushToken: u.WebPushToken != "",
		})
	}
	return resp
}

func ToGroupResponses(groups []*domain.Group) []*GroupResponse {
	resps := make([]*GroupResponse, len(groups))
	for i, g := range groups {
		resps[i] = ToGroupResponse(g)
	}
	return resps
}

// RecurrenceSettingsDTO は繰り返しの設定DTOである．
type RecurrenceSettingsDTO struct {
	Type        string      `json:"type"`
	CustomDates []time.Time `json:"custom_dates"`
}

// TaskUserProgressResponse は課題進捗状況のDTOである．
type TaskUserProgressResponse struct {
	TaskID      string    `json:"task_id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	IsCompleted bool      `json:"is_completed"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskResponse はクライアントへ返却する課題情報のDTOである．
type TaskResponse struct {
	ID               string                      `json:"id"`
	GroupID          string                      `json:"group_id"`
	Source           string                      `json:"source"`
	ExternalID       string                      `json:"external_id"`
	CreatorID        string                      `json:"creator_id"`
	Title            string                      `json:"title"`
	Subject          string                      `json:"subject"`
	Deadline         time.Time                   `json:"deadline"`
	IsLMSDeadlineSet bool                        `json:"is_lms_deadline_set"`
	LMSUpdateTime    time.Time                   `json:"lms_update_time"`
	Recurrence       RecurrenceSettingsDTO       `json:"recurrence"`
	UserProgress     []*TaskUserProgressResponse `json:"user_progress"`
}

func ToTaskResponse(t *domain.Task) *TaskResponse {
	if t == nil {
		return nil
	}
	resp := &TaskResponse{
		ID:               t.ID,
		GroupID:          t.GroupID,
		Source:           t.Source,
		ExternalID:       t.ExternalID,
		CreatorID:        t.CreatorID,
		Title:            t.Title,
		Subject:          t.Subject,
		Deadline:         t.Deadline,
		IsLMSDeadlineSet: t.IsLMSDeadlineSet,
		LMSUpdateTime:    t.LMSUpdateTime,
		Recurrence: RecurrenceSettingsDTO{
			Type:        t.Recurrence.Type,
			CustomDates: t.Recurrence.CustomDates,
		},
	}
	for _, p := range t.UserProgress {
		resp.UserProgress = append(resp.UserProgress, &TaskUserProgressResponse{
			TaskID:      p.TaskID,
			UserID:      p.UserID,
			IsCompleted: p.IsCompleted,
			UpdatedAt:   p.UpdatedAt,
		})
	}
	return resp
}

func ToTaskResponses(tasks []*domain.Task) []*TaskResponse {
	resps := make([]*TaskResponse, len(tasks))
	for i, t := range tasks {
		resps[i] = ToTaskResponse(t)
	}
	return resps
}

// TaskRequest は課題作成・更新リクエスト用のDTOである．
type TaskRequest struct {
	ID               string                      `json:"id"`
	GroupID          string                      `json:"group_id"`
	Source           string                      `json:"source"`
	ExternalID       string                      `json:"external_id"`
	CreatorID        string                      `json:"creator_id"`
	Title            string                      `json:"title"`
	Subject          string                      `json:"subject"`
	Deadline         time.Time                   `json:"deadline"`
	IsLMSDeadlineSet bool                        `json:"is_lms_deadline_set"`
	Recurrence       RecurrenceSettingsDTO       `json:"recurrence"`
	UserProgress     []*TaskUserProgressResponse `json:"user_progress"`
}

func (req *TaskRequest) ToDomain() *domain.Task {
	t := &domain.Task{
		ID:               req.ID,
		GroupID:          req.GroupID,
		Source:           req.Source,
		ExternalID:       req.ExternalID,
		CreatorID:        req.CreatorID,
		Title:            req.Title,
		Subject:          req.Subject,
		Deadline:         req.Deadline,
		IsLMSDeadlineSet: req.IsLMSDeadlineSet,
		Recurrence: domain.RecurrenceSettings{
			Type:        req.Recurrence.Type,
			CustomDates: req.Recurrence.CustomDates,
		},
	}
	for _, p := range req.UserProgress {
		t.UserProgress = append(t.UserProgress, &domain.TaskUserProgress{
			TaskID:      p.TaskID,
			UserID:      p.UserID,
			IsCompleted: p.IsCompleted,
			UpdatedAt:   p.UpdatedAt,
		})
	}
	return t
}

// WakeupCheckResponse はクライアントへ返却する起床確認DTOである．
type WakeupCheckResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	GroupID      string    `json:"group_id"`
	TargetTime   time.Time `json:"target_time"`
	GraceMinutes int       `json:"grace_minutes"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func ToWakeupResponse(w *domain.WakeupCheck) *WakeupCheckResponse {
	if w == nil {
		return nil
	}
	return &WakeupCheckResponse{
		ID:           w.ID,
		UserID:       w.UserID,
		GroupID:      w.GroupID,
		TargetTime:   w.TargetTime,
		GraceMinutes: w.GraceMinutes,
		Status:       w.Status,
		CreatedAt:    w.CreatedAt,
	}
}

func ToWakeupResponses(checks []*domain.WakeupCheck) []*WakeupCheckResponse {
	resps := make([]*WakeupCheckResponse, len(checks))
	for i, c := range checks {
		resps[i] = ToWakeupResponse(c)
	}
	return resps
}

// NotificationLogResponse は通知履歴のDTOである．
type NotificationLogResponse struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func ToNotificationLogResponse(l *domain.NotificationLog) *NotificationLogResponse {
	if l == nil {
		return nil
	}
	return &NotificationLogResponse{
		ID:        l.ID,
		GroupID:   l.GroupID,
		UserID:    l.UserID,
		Type:      l.Type,
		Message:   l.Message,
		CreatedAt: l.CreatedAt,
	}
}

func ToNotificationLogResponses(logs []*domain.NotificationLog) []*NotificationLogResponse {
	resps := make([]*NotificationLogResponse, len(logs))
	for i, l := range logs {
		resps[i] = ToNotificationLogResponse(l)
	}
	return resps
}
