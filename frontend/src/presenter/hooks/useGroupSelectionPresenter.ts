import { useState, useCallback } from 'react';
import type { Group } from '../../domain/models';
import { groupUsecase } from '../../di/container';
import { handle } from '../../utils/helpers';
import { AxiosError } from 'axios';

/**
 * グループ選択画面のプレゼンターフック（状態管理・ユースケース呼び出し）
 */
export const useGroupSelectionPresenter = (userId: string) => {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchGroups = useCallback(async () => {
    if (!userId) return;
    setError(null);
    const [data, err] = await handle(groupUsecase.fetchUserGroups(userId));
    if (err) {
      console.error("useGroupSelectionPresenter: fetchGroups failed", err.message);
      setError("グループの取得に失敗しました。");
      setLoading(false);
      return;
    }
    setGroups(data || []);
    setLoading(false);
  }, [userId]);

  const createGroup = useCallback(async (name: string): Promise<Group | null> => {
    if (!name.trim()) return null;
    setLoading(true);
    const [newGroup, err] = await handle(groupUsecase.createGroup(name, userId));
    setLoading(false);
    if (err) {
      const axiosErr = err as AxiosError<{ error: string }>;
      alert(`エラー：${axiosErr.response?.data?.error || "部屋の作成に失敗しました。"}`);
      return null;
    }
    setGroups(prev => [...prev, newGroup]);
    return newGroup;
  }, [userId]);

  const joinGroup = useCallback(async (inviteCode: string): Promise<Group | null> => {
    if (!inviteCode.trim()) return null;
    setLoading(true);
    const [joinedGroup, err] = await handle(groupUsecase.joinGroup(inviteCode, userId));
    setLoading(false);
    if (err) {
      const axiosErr = err as AxiosError<{ error: string }>;
      alert(`エラー：${axiosErr.response?.data?.error || "部屋への参加に失敗しました。招待コードを確認してください。"}`);
      return null;
    }
    return joinedGroup;
  }, [userId]);

  return {
    groups,
    loading,
    error,
    fetchGroups,
    createGroup,
    joinGroup,
  };
};
