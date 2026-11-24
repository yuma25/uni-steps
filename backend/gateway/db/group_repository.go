package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

type groupRepository struct {
	db *sql.DB
}

// NewGroupRepository は groupRepository の新しいインスタンスを生成する．
func NewGroupRepository(db *sql.DB) port.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

// Save はグループ情報を保存または更新する（Upsert）．
func (r *groupRepository) Save(ctx context.Context, group *domain.Group) error {
	intervalsJSON, err := json.Marshal(group.RemindIntervals)
	if err != nil {
		intervalsJSON = []byte("[]")
	}

	query := `
		INSERT INTO groups (
			id, name, owner_id, line_channel_token, line_group_id,
			last_synced_at, lms_last_updated_at, invite_code,
			remind_intervals, ai_character, summary_morning_time, summary_evening_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			owner_id = EXCLUDED.owner_id,
			line_channel_token = EXCLUDED.line_channel_token,
			line_group_id = EXCLUDED.line_group_id,
			last_synced_at = EXCLUDED.last_synced_at,
			lms_last_updated_at = EXCLUDED.lms_last_updated_at,
			invite_code = EXCLUDED.invite_code,
			remind_intervals = EXCLUDED.remind_intervals,
			ai_character = EXCLUDED.ai_character,
			summary_morning_time = EXCLUDED.summary_morning_time,
			summary_evening_time = EXCLUDED.summary_evening_time;
	`

	_, err = r.db.ExecContext(ctx, query,
		group.ID,
		group.Name,
		group.OwnerID,
		group.LineChannelToken,
		group.LineGroupID,
		group.LastSyncedAt,
		group.LMSLastUpdatedAt,
		group.InviteCode,
		intervalsJSON,
		group.AICharacter,
		group.SummaryMorningTime,
		group.SummaryEveningTime,
	)
	return err
}

func (r *groupRepository) FindAllGroups(ctx context.Context) ([]*domain.Group, error) {
	query := `
		SELECT id, name, owner_id, line_channel_token, line_group_id,
		       last_synced_at, lms_last_updated_at, invite_code,
		       remind_intervals, ai_character, summary_morning_time, summary_evening_time
		FROM groups;
	`
	return r.queryGroups(ctx, query)
}

func (r *groupRepository) FindByID(ctx context.Context, id string) (*domain.Group, error) {
	query := `
		SELECT id, name, owner_id, line_channel_token, line_group_id,
		       last_synced_at, lms_last_updated_at, invite_code,
		       remind_intervals, ai_character, summary_morning_time, summary_evening_time
		FROM groups
		WHERE id = $1;
	`
	groups, err := r.queryGroups(ctx, query, id)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	return groups[0], nil
}

func (r *groupRepository) FindByInviteCode(ctx context.Context, code string) (*domain.Group, error) {
	query := `
		SELECT id, name, owner_id, line_channel_token, line_group_id,
		       last_synced_at, lms_last_updated_at, invite_code,
		       remind_intervals, ai_character, summary_morning_time, summary_evening_time
		FROM groups
		WHERE invite_code = $1;
	`
	groups, err := r.queryGroups(ctx, query, code)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	return groups[0], nil
}

func (r *groupRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Group, error) {
	query := `
		SELECT g.id, g.name, g.owner_id, g.line_channel_token, g.line_group_id,
		       g.last_synced_at, g.lms_last_updated_at, g.invite_code,
		       g.remind_intervals, g.ai_character, g.summary_morning_time, g.summary_evening_time
		FROM groups g
		JOIN user_groups ug ON ug.group_id = g.id
		WHERE ug.user_id = $1;
	`
	return r.queryGroups(ctx, query, userID)
}

func (r *groupRepository) RemoveUser(ctx context.Context, groupID string, userID string) error {
	query := `DELETE FROM user_groups WHERE group_id = $1 AND user_id = $2;`
	_, err := r.db.ExecContext(ctx, query, groupID, userID)
	return err
}

func (r *groupRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_groups WHERE group_id = $1;`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = $1;`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *groupRepository) AddUserToGroup(ctx context.Context, groupID string, userID string) error {
	// グループの存在確認
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1);`, groupID).Scan(&exists); err != nil || !exists {
		if !exists {
			return errors.New("指定されたグループが見つかりません")
		}
		return err
	}
	// ユーザーの存在確認
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);`, userID).Scan(&exists); err != nil || !exists {
		if !exists {
			return errors.New("指定されたユーザーが見つかりません")
		}
		return err
	}

	query := `
		INSERT INTO user_groups (user_id, group_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`
	_, err := r.db.ExecContext(ctx, query, userID, groupID)
	return err
}

func (r *groupRepository) queryGroups(ctx context.Context, query string, args ...any) ([]*domain.Group, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var g domain.Group
		var intervalsJSON []byte
		var lastSynced, lmsUpdated sql.NullTime

		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.OwnerID,
			&g.LineChannelToken,
			&g.LineGroupID,
			&lastSynced,
			&lmsUpdated,
			&g.InviteCode,
			&intervalsJSON,
			&g.AICharacter,
			&g.SummaryMorningTime,
			&g.SummaryEveningTime,
		); err != nil {
			return nil, err
		}
		if lastSynced.Valid {
			g.LastSyncedAt = lastSynced.Time
		}
		if lmsUpdated.Valid {
			g.LMSLastUpdatedAt = lmsUpdated.Time
		}
		if len(intervalsJSON) > 0 {
			_ = json.Unmarshal(intervalsJSON, &g.RemindIntervals)
		}
		groups = append(groups, &g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 各グループのメンバー一覧（Users）を取得
	for _, g := range groups {
		users, err := r.fetchUsersForGroup(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		g.Users = users
	}

	return groups, nil
}

func (r *groupRepository) fetchUsersForGroup(ctx context.Context, groupID string) ([]*domain.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.web_push_token
		FROM users u
		JOIN user_groups ug ON ug.user_id = u.id
		WHERE ug.group_id = $1;
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Email,
			&u.WebPushToken,
		); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, rows.Err()
}
