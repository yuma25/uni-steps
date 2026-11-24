package infra

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDB は PostgreSQL (Supabase) への接続を初期化し，テーブル作成とマイグレーションを実行する．
func InitDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL が設定されていない．.env ファイルを確認すること")
	}

	if !strings.Contains(dbURL, "TimeZone=") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&TimeZone=Asia/Tokyo"
		} else {
			dbURL += "?TimeZone=Asia/Tokyo"
		}
	}

	// Supabase (PgBouncer: ポート6543) の Transaction Pooling で
	// "prepared statement already exists" (42P05) エラーが発生するのを防止する設定
	if !strings.Contains(dbURL, "default_query_exec_mode=") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&default_query_exec_mode=simple_protocol"
		} else {
			dbURL += "?default_query_exec_mode=simple_protocol"
		}
	}

	log.Println("データベースへの接続を開始中（最大5秒待機）...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("データベース接続設定に失敗した: %w", err)
	}

	// 接続プールの基本設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("データベースへの Ping に失敗した: %w", err)
	}

	// テーブル初期化（生SQL DDL）
	if err := migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("テーブルの初期化に失敗した: %w", err)
	}

	log.Println("データベース接続およびテーブルマイグレーションが完了した")
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			email VARCHAR(255) UNIQUE NOT NULL DEFAULT '',
			web_push_token TEXT NOT NULL DEFAULT '',
			google_access_token TEXT NOT NULL DEFAULT '',
			google_refresh_token TEXT NOT NULL DEFAULT '',
			google_token_expiry TIMESTAMPTZ
		);`,
		`ALTER TABLE users DROP COLUMN IF EXISTS last_check_in_at;`,
		`CREATE TABLE IF NOT EXISTS groups (
			id VARCHAR(255) PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			owner_id VARCHAR(255) NOT NULL DEFAULT '',
			line_channel_token TEXT NOT NULL DEFAULT '',
			line_group_id TEXT NOT NULL DEFAULT '',
			last_synced_at TIMESTAMPTZ,
			lms_last_updated_at TIMESTAMPTZ,
			invite_code VARCHAR(255) UNIQUE NOT NULL DEFAULT '',
			remind_intervals JSONB NOT NULL DEFAULT '[]',
			ai_character TEXT NOT NULL DEFAULT 'default',
			summary_morning_time TEXT NOT NULL DEFAULT '',
			summary_evening_time TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS user_groups (
			user_id VARCHAR(255) NOT NULL,
			group_id VARCHAR(255) NOT NULL,
			PRIMARY KEY (user_id, group_id)
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(255) PRIMARY KEY,
			group_id VARCHAR(255) NOT NULL DEFAULT '',
			source VARCHAR(50) NOT NULL DEFAULT 'manual',
			external_id VARCHAR(255) NOT NULL DEFAULT '',
			creator_id VARCHAR(255) NOT NULL DEFAULT '',
			title TEXT NOT NULL DEFAULT '',
			subject TEXT NOT NULL DEFAULT '',
			deadline TIMESTAMPTZ,
			is_lms_deadline_set BOOLEAN NOT NULL DEFAULT FALSE,
			lms_update_time TIMESTAMPTZ,
			recurrence JSONB NOT NULL DEFAULT '{}'
		);`,
		`ALTER TABLE tasks DROP COLUMN IF EXISTS raw_text;`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS subject TEXT NOT NULL DEFAULT '';`,
		`CREATE TABLE IF NOT EXISTS task_user_progresses (
			task_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			is_completed BOOLEAN NOT NULL DEFAULT FALSE,
			updated_at TIMESTAMPTZ,
			PRIMARY KEY (task_id, user_id)
		);`,
		`ALTER TABLE task_user_progresses DROP COLUMN IF EXISTS user_name;`,
		`CREATE TABLE IF NOT EXISTS wakeup_checks (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			group_id VARCHAR(255) NOT NULL,
			target_time TIMESTAMPTZ NOT NULL,
			grace_minutes INT NOT NULL DEFAULT 0,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ
		);`,
		`CREATE TABLE IF NOT EXISTS notification_logs (
			id VARCHAR(255) PRIMARY KEY,
			group_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL DEFAULT '',
			type VARCHAR(50) NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ
		);`,
		// 複合ユニークインデックス等の適用
		`DROP INDEX IF EXISTS idx_tasks_external_id;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_group_external ON tasks (group_id, external_id) WHERE external_id <> '';`,
		`CREATE INDEX IF NOT EXISTS idx_notification_logs_group_id ON notification_logs (group_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_group_id ON tasks (group_id);`,
		`CREATE INDEX IF NOT EXISTS idx_user_groups_group_id ON user_groups (group_id);`,
		`CREATE INDEX IF NOT EXISTS idx_task_user_progresses_task_id ON task_user_progresses (task_id);`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("クエリ実行エラー [%s]: %w", query, err)
		}
	}

	return nil
}
