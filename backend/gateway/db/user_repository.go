package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository は userRepository の新しいインスタンスを生成する．
func NewUserRepository(db *sql.DB) port.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Save はユーザーをデータベースに保存する（Upsert）．
func (r *userRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, name, email, web_push_token, google_access_token,
			google_refresh_token, google_token_expiry
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			web_push_token = EXCLUDED.web_push_token,
			google_access_token = EXCLUDED.google_access_token,
			google_refresh_token = EXCLUDED.google_refresh_token,
			google_token_expiry = EXCLUDED.google_token_expiry;
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.WebPushToken,
		user.GoogleAccessToken,
		user.GoogleRefreshToken,
		user.GoogleTokenExpiry,
	)
	return err
}

// UpdateWebPushToken はユーザーの Web Push トークンを明示的に更新する．
func (r *userRepository) UpdateWebPushToken(ctx context.Context, userID string, token string) error {
	query := `UPDATE users SET web_push_token = $1 WHERE id = $2;`
	_, err := r.db.ExecContext(ctx, query, token, userID)
	return err
}

// FindByID は指定された ID のユーザーを取得する．
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, name, email, web_push_token, google_access_token,
		       google_refresh_token, google_token_expiry
		FROM users
		WHERE id = $1;
	`
	var user domain.User
	var tokenExpiry sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.WebPushToken,
		&user.GoogleAccessToken,
		&user.GoogleRefreshToken,
		&tokenExpiry,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if tokenExpiry.Valid {
		user.GoogleTokenExpiry = tokenExpiry.Time
	}

	// 所属グループを取得
	groups, err := r.fetchGroupsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Groups = groups

	return &user, nil
}

// FindByEmail は指定されたメールアドレスのユーザーを取得する．
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, web_push_token, google_access_token,
		       google_refresh_token, google_token_expiry
		FROM users
		WHERE email = $1;
	`
	var user domain.User
	var tokenExpiry sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.WebPushToken,
		&user.GoogleAccessToken,
		&user.GoogleRefreshToken,
		&tokenExpiry,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if tokenExpiry.Valid {
		user.GoogleTokenExpiry = tokenExpiry.Time
	}

	// 所属グループを取得
	groups, err := r.fetchGroupsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Groups = groups

	return &user, nil
}

func (r *userRepository) fetchGroupsForUser(ctx context.Context, userID string) ([]*domain.Group, error) {
	query := `
		SELECT g.id, g.name, g.owner_id, g.invite_code, g.remind_intervals,
		       g.ai_character, g.summary_morning_time, g.summary_evening_time
		FROM groups g
		JOIN user_groups ug ON ug.group_id = g.id
		WHERE ug.user_id = $1;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var g domain.Group
		var intervalsJSON []byte
		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.OwnerID,
			&g.InviteCode,
			&intervalsJSON,
			&g.AICharacter,
			&g.SummaryMorningTime,
			&g.SummaryEveningTime,
		); err != nil {
			return nil, err
		}
		if len(intervalsJSON) > 0 {
			_ = json.Unmarshal(intervalsJSON, &g.RemindIntervals)
		}
		groups = append(groups, &g)
	}

	return groups, rows.Err()
}
