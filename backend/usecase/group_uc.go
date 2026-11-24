package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

// GroupUsecase はグループ（部屋）の作成や参加に関するビジネスロジックを担当する構造体．
type GroupUsecase struct {
	groupRepo port.GroupRepository           // グループデータの永続化を担うリポジトリ．
	userRepo  port.UserRepository            // ユーザー情報を確認するためのリポジトリ．
	logRepo   port.NotificationLogRepository // 通知履歴を取得するためのリポジトリ．
	summaryUC *SummaryUsecase                // 定期サマリーのタイマー予約を管理するユースケース．
}

// NewGroupUsecase は GroupUsecase の新しいインスタンスを生成する．
func NewGroupUsecase(gr port.GroupRepository, ur port.UserRepository, lr port.NotificationLogRepository, su *SummaryUsecase) *GroupUsecase {
	return &GroupUsecase{
		groupRepo: gr,
		userRepo:  ur,
		logRepo:   lr,
		summaryUC: su,
	}
}

// CreateGroup は新しいグループを作成し，作成者をオーナーとして登録する．
func (uc *GroupUsecase) CreateGroup(ctx context.Context, name string, ownerID string) (*domain.Group, error) {
	// オーナーとなるユーザーが存在するか確認する．
	user, err := uc.userRepo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("オーナー情報の取得に失敗した： %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("オーナーとなるユーザーが見つからない")
	}

	// 新しいグループ構造体を作成する．
	inviteCode := uuid.New().String()[:8]

	group := &domain.Group{
		ID:                 uuid.New().String(),
		Name:               name,
		OwnerID:            ownerID,
		InviteCode:         inviteCode,
		RemindIntervals:    []int{1440, 180, 60},
		AICharacter:        domain.AICharacterDefault,
		SummaryMorningTime: "06:00",
		SummaryEveningTime: "22:00",
		Users:              []*domain.User{},
	}

	// まずデータベースに「部屋」だけを保存する．
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("グループの保存に失敗した： %w", err)
	}

	// 保存された部屋に対して，オーナーを所属させる．
	if err := uc.groupRepo.AddUserToGroup(ctx, group.ID, user.ID); err != nil {
		return nil, fmt.Errorf("メンバーの紐付けに失敗した： %w", err)
	}

	// ④ 副作用の実行: DB保存が成功した後にだけ、外部連携を行う（定期サマリータイマーを予約）
	if uc.summaryUC != nil {
		uc.summaryUC.ScheduleGroupSummary(ctx, group)
	}

	return group, nil
}

// JoinGroupByInviteCode は招待コードを用いて既存のグループに参加する．
func (uc *GroupUsecase) JoinGroupByInviteCode(ctx context.Context, code string, userID string) (*domain.Group, error) {
	// 招待コードに該当するグループを探す．
	group, err := uc.groupRepo.FindByInviteCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("招待コードの検証に失敗した： %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("無効な招待コードである")
	}

	// 参加するユーザーが存在するか確認する．
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("ユーザーが見つからない")
	}

	// 既に参加していないか確認する．
	for _, u := range group.Users {
		if u.ID == userID {
			return group, nil // 既に参加済みの場合はそのまま成功とする
		}
	}

	// ユーザーをグループに追加して保存する．
	if err := uc.groupRepo.AddUserToGroup(ctx, group.ID, user.ID); err != nil {
		return nil, fmt.Errorf("グループへの参加に失敗した： %w", err)
	}

	return group, nil
}

// ListUserGroups は指定されたユーザーが所属しているグループの一覧を取得する．
func (uc *GroupUsecase) ListUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	groups, err := uc.groupRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("所属グループ一覧の取得に失敗した： %w", err)
	}
	return groups, nil
}

// GetGroup は指定されたIDのグループ詳細を取得する．
func (uc *GroupUsecase) GetGroup(ctx context.Context, groupID string) (*domain.Group, error) {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("グループ情報の取得に失敗した： %w", err)
	}
	return group, nil
}

// UpdateSettings は部屋の設定を更新する．オーナーのみ許可する．
func (uc *GroupUsecase) UpdateSettings(ctx context.Context, groupID string, userID string, name string, intervals []int, aiCharacter string, lineToken string, lineGroupID string, morningTime string, eveningTime string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	if group.OwnerID != userID {
		return fmt.Errorf("この部屋の設定を変更する権限がない")
	}

	if name != "" {
		group.Name = name
	}
	group.RemindIntervals = intervals
	group.AICharacter = aiCharacter
	group.LineChannelToken = lineToken
	group.LineGroupID = lineGroupID
	group.SummaryMorningTime = morningTime
	group.SummaryEveningTime = eveningTime

	// ③ データベースへ保存
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return err
	}

	// ④ 副作用の実行: DB保存が成功した後にだけ、外部連携を行う（タイマーを破棄して再予約）
	if uc.summaryUC != nil {
		uc.summaryUC.CancelGroupSummary(ctx, group.ID)
		uc.summaryUC.ScheduleGroupSummary(ctx, group)
	}

	return nil
}

// LeaveGroup はユーザーを部屋から退出させる．
func (uc *GroupUsecase) LeaveGroup(ctx context.Context, groupID string, userID string) error {
	// 部屋が存在するか確認する．
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	// オーナーが退出する場合は，事前に権譲譲が必要であることを伝える．
	if group.OwnerID == userID && len(group.Users) > 1 {
		return fmt.Errorf("オーナーは退出前に次のオーナーを指名する必要がある")
	}

	// リポジトリ経由で紐付けを削除する．
	return uc.groupRepo.RemoveUser(ctx, groupID, userID)
}

// TransferOwnership は部屋のオーナー権限を別のユーザーに譲渡する．
func (uc *GroupUsecase) TransferOwnership(ctx context.Context, groupID string, currentOwnerID string, newOwnerID string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	// 現在のオーナーであるか確認する．
	if group.OwnerID != currentOwnerID {
		return fmt.Errorf("オーナー権限の譲渡は現在のオーナーのみ可能である")
	}

	// 新しいオーナーがグループに所属しているか確認する．
	found := false
	for _, u := range group.Users {
		if u.ID == newOwnerID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("新しいオーナーはグループのメンバーである必要がある")
	}

	group.OwnerID = newOwnerID
	return uc.groupRepo.Save(ctx, group)
}

// DeleteGroup は部屋を完全に削除する．オーナーのみ許可する．
func (uc *GroupUsecase) DeleteGroup(ctx context.Context, groupID string, userID string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	if group.OwnerID != userID {
		return fmt.Errorf("この部屋を削除する権限がない（オーナーのみ可能）")
	}

	return uc.groupRepo.Delete(ctx, groupID)
}

// GetNotificationLogs は指定されたグループの通知履歴を取得する．
func (uc *GroupUsecase) GetNotificationLogs(ctx context.Context, groupID string, limit int) ([]*domain.NotificationLog, error) {
	return uc.logRepo.FindByGroupID(ctx, groupID, limit)
}
