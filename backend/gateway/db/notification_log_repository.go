package db

import (
	"context"
	"database/sql"

	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

type notificationLogRepository struct {
	db *sql.DB
}

// NewNotificationLogRepository は新しいインスタンスを生成する．
func NewNotificationLogRepository(db *sql.DB) port.NotificationLogRepository {
	return &notificationLogRepository{
		db: db,
	}
}

// Save は新しい通知ログを保存する．
func (r *notificationLogRepository) Save(ctx context.Context, log *domain.NotificationLog) error {
	query := `
		INSERT INTO notification_logs (id, group_id, user_id, type, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.GroupID,
		log.UserID,
		log.Type,
		log.Message,
		log.CreatedAt,
	)
	return err
}

// FindByGroupID は指定された部屋の過去の通知ログを新しい順に取得する．
func (r *notificationLogRepository) FindByGroupID(ctx context.Context, groupID string, limit int) ([]*domain.NotificationLog, error) {
	query := `
		SELECT id, group_id, user_id, type, message, created_at
		FROM notification_logs
		WHERE group_id = $1
		ORDER BY created_at DESC
		LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.NotificationLog
	for rows.Next() {
		var l domain.NotificationLog
		var createdAt sql.NullTime

		if err := rows.Scan(
			&l.ID,
			&l.GroupID,
			&l.UserID,
			&l.Type,
			&l.Message,
			&createdAt,
		); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			l.CreatedAt = createdAt.Time
		}
		logs = append(logs, &l)
	}

	return logs, rows.Err()
}
