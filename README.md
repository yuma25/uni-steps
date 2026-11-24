# Uni-Steps

AIによる課題解析・リマインドと、仲間との起床見守りを統合した、次世代のライフリズム支援プラットフォームです。
「コストゼロ」で「プロフェッショナル品質」のサービスを実現するための、クリーンアーキテクチャに基づいたフルスタック・アプリケーションです。

![Architecture](./docs/images/architecture.png)

---

## コア・コンセプト

**「一人じゃないから、一歩踏み出せる」**
Uni-Stepsは、単なるタスク管理ツールではありません。グループ内での予定共有にAIが介入し、最適なタイミングとメッセージでリマインドを行うことで、タスクの完遂と良好な生活リズムを「仲間と一緒に」作り出すことを目的としています。

---

## 主要機能

- **AI スマート・リマインド**: Google Gemini APIを活用し、課題の内容や締切に応じたメッセージを自動生成。AIの性格（教官・幼馴染・執事など）を部屋ごとにカスタマイズ可能。
- **朝・夜のデイリーサマリー自動配信**: 常駐スケジューラーが設定時刻にグループ全体の課題進捗と未完了メンバーをAIで要約し、LINE通知。
- **LMS フル同期**: Google Classroom等と完全連携。Goroutineを用いた並列取得により、複数コースの課題を瞬時に同期。
- **クリティカル起床確認**: 絶対に外せない予定がある朝、チェックインがない場合にグループ全員へSOS通知を自動配信。仲間のサポートで寝坊を防止。
- **ハイブリッド通知**: LINE（グループ共有）と Web Push（個人宛PWA通知）を組み合わせた通知システム。失効したトークンの自己修復機能を搭載。

---

## テックスタック


| カテゴリ           | 採用技術                                               |
| :------------------- | :------------------------------------------------------- |
| **Frontend**       | React 19 / TypeScript / Vite                           |
| **Backend**        | Go 1.22 / Echo v4 / pgx (生SQL)                        |
| **Database**       | Supabase (PostgreSQL) / PgBouncer 互換 Simple Protocol |
| **Intelligence**   | Google Gemini API (2.5 Flash / Lite)                   |
| **Messaging**      | LINE Messaging API / Web Push (VAPID)                  |
| **Infrastructure** | Vercel / Render                                        |

---

## アーキテクチャ

バックエンド・フロントエンドの双方がクリーンアーキテクチャの原則に厳格に従い、ビジネスロジックを外部技術（フレームワーク、DB、通信ライブラリ）から完全に独立させています。

### バックエンド（Clean Architecture）

1. **Domain**: エンティティおよび不変ルール（ビジネスロジック）を保持。外部ライブラリへの依存ゼロ。
2. **Port**: UseCase が要求するリポジトリや外部サービスの抽象インターフェース。
3. **UseCase**: 各業務シナリオの手順を純粋なロジックとして制御。
4. **Gateway**: Port を実装し、生SQL（pgx）によるデータベースアクセスや外部API連携を実行。
5. **Handler**: Echo の HTTP コンテキストを解釈し、UseCase へ橋渡し。
6. **DI**: 全インスタンスを生成・注入して配線。

### フロントエンド（8層 Clean Architecture）

```
src/
├── domain/       第1層: エンティティ型・不変ルール（React/通信非依存）
├── port/         第2層: API通信の抽象インターフェース
├── usecase/      第3層: 純粋なTypeScriptによるユースケース手順制御
├── infra/        第4層/第5層: Portの具象Gateway実装およびHTTP/ブラウザインフラ
├── presenter/    第6層: Custom Hooksによる画面状態管理とViewModel変換
├── view/         第7層: 純粋JSX表示コンポーネント・画面（pages, components, context, hooks）
├── di/           第8層: インスタンス生成と依存注入（DI Container）
├── App.tsx       第8層: 全体ルーティング
└── main.tsx      第8層: エントリーポイント
```

---

## ディレクトリ構成

- `backend/`: GoによるAPIサーバー
  - `domain/`: システムの中核ルールとエンティティ
  - `port/`: 抽象インターフェース定義
  - `usecase/`: 業務ユースケース
  - `gateway/`: DB（生SQL）・AI・LINE・WebPushの具象実装
  - `handler/`: HTTPリクエストハンドラー
  - `di/`: 依存注入コンテナ
  - `infra/`: DB接続・ミドルウェア基盤
- `frontend/`: React + TypeScript によるフロントエンド
  - `src/domain/`: ドメインモデル
  - `src/port/`: ポートインターフェース
  - `src/usecase/`: ユースケース実装
  - `src/infra/`: APIクライアント・Gateway実装
  - `src/presenter/`: Presenterカスタムフック群
  - `src/view/`: UI表現層（ページ・コンポーネント・コンテキスト）
  - `src/di/`: DIコンテナ
- `docs/`: 画像リソース等

---

*Uni-Steps: Your life rhythm partner.*
