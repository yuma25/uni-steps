package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) port.TaskRepository {
	return &taskRepository{db: db}
}

// Save はタスクを保存または更新し，その進捗状況（UserProgress）を同期する．
func (r *taskRepository) Save(ctx context.Context, task *domain.Task) error {
	recurrenceJSON, err := json.Marshal(task.Recurrence)
	if err != nil {
		recurrenceJSON = []byte("{}")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	taskQuery := `
		INSERT INTO tasks (
			id, group_id, source, external_id, creator_id,
			title, subject, deadline, is_lms_deadline_set, lms_update_time, recurrence
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			group_id = EXCLUDED.group_id,
			source = EXCLUDED.source,
			external_id = EXCLUDED.external_id,
			creator_id = EXCLUDED.creator_id,
			title = EXCLUDED.title,
			subject = EXCLUDED.subject,
			deadline = EXCLUDED.deadline,
			is_lms_deadline_set = EXCLUDED.is_lms_deadline_set,
			lms_update_time = EXCLUDED.lms_update_time,
			recurrence = EXCLUDED.recurrence;
	`

	_, err = tx.ExecContext(ctx, taskQuery,
		task.ID,
		task.GroupID,
		task.Source,
		task.ExternalID,
		task.CreatorID,
		task.Title,
		task.Subject,
		task.Deadline,
		task.IsLMSDeadlineSet,
		task.LMSUpdateTime,
		recurrenceJSON,
	)
	if err != nil {
		return err
	}

	// 2. UserProgress 同期
	if _, err := tx.ExecContext(ctx, `DELETE FROM task_user_progresses WHERE task_id = $1;`, task.ID); err != nil {
		return err
	}

	if len(task.UserProgress) > 0 {
		progressQuery := `
			INSERT INTO task_user_progresses (task_id, user_id, is_completed, updated_at)
			VALUES ($1, $2, $3, $4);
		`
		for _, p := range task.UserProgress {
			p.TaskID = task.ID
			if _, err := tx.ExecContext(ctx, progressQuery, p.TaskID, p.UserID, p.IsCompleted, p.UpdatedAt); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, group_id, source, external_id, creator_id,
		       title, subject, deadline, is_lms_deadline_set, lms_update_time, recurrence
		FROM tasks
		WHERE id = $1;
	`
	tasks, err := r.queryTasks(ctx, query, id)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks[0], nil
}

func (r *taskRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.Task, error) {
	query := `
		SELECT id, group_id, source, external_id, creator_id,
		       title, subject, deadline, is_lms_deadline_set, lms_update_time, recurrence
		FROM tasks
		WHERE external_id = $1;
	`
	tasks, err := r.queryTasks(ctx, query, externalID)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks[0], nil
}

func (r *taskRepository) FindByGroupID(ctx context.Context, groupID string) ([]*domain.Task, error) {
	query := `
		SELECT id, group_id, source, external_id, creator_id,
		       title, subject, deadline, is_lms_deadline_set, lms_update_time, recurrence
		FROM tasks
		WHERE group_id = $1
		ORDER BY deadline ASC;
	`
	return r.queryTasks(ctx, query, groupID)
}

func (r *taskRepository) FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*domain.Task, error) {
	query := `
		SELECT id, group_id, source, external_id, creator_id,
		       title, subject, deadline, is_lms_deadline_set, lms_update_time, recurrence
		FROM tasks
		WHERE deadline <= $1
		ORDER BY deadline ASC;
	`
	return r.queryTasks(ctx, query, until)
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM task_user_progresses WHERE task_id = $1;`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1;`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *taskRepository) queryTasks(ctx context.Context, query string, args ...any) ([]*domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var t domain.Task
		var recurrenceJSON []byte
		var deadline, lmsUpdate sql.NullTime

		if err := rows.Scan(
			&t.ID,
			&t.GroupID,
			&t.Source,
			&t.ExternalID,
			&t.CreatorID,
			&t.Title,
			&t.Subject,
			&deadline,
			&t.IsLMSDeadlineSet,
			&lmsUpdate,
			&recurrenceJSON,
		); err != nil {
			return nil, err
		}
		if deadline.Valid {
			t.Deadline = deadline.Time
		}
		if lmsUpdate.Valid {
			t.LMSUpdateTime = lmsUpdate.Time
		}
		if len(recurrenceJSON) > 0 {
			_ = json.Unmarshal(recurrenceJSON, &t.Recurrence)
		}
		tasks = append(tasks, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 各タスクの UserProgress を取得
	for _, t := range tasks {
		progresses, err := r.fetchProgressForTask(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		t.UserProgress = progresses
	}

	return tasks, nil
}

func (r *taskRepository) fetchProgressForTask(ctx context.Context, taskID string) ([]*domain.TaskUserProgress, error) {
	query := `
		SELECT task_id, user_id, is_completed, updated_at
		FROM task_user_progresses
		WHERE task_id = $1;
	`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progresses []*domain.TaskUserProgress
	for rows.Next() {
		var p domain.TaskUserProgress
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&p.TaskID,
			&p.UserID,
			&p.IsCompleted,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if updatedAt.Valid {
			p.UpdatedAt = updatedAt.Time
		}
		progresses = append(progresses, &p)
	}

	return progresses, rows.Err()
}
