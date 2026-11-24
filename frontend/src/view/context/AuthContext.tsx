import { useState, useEffect, useCallback, type ReactNode } from 'react';
import { useSearchParams } from 'react-router-dom';
import { userUsecase } from '../../di/container';
import type { User } from '../../domain/models';
import { AuthContext } from './AuthContextInstance';
import { handle } from '../../utils/helpers';

const STORAGE_KEY_USER_ID = 'uni_steps_user_id';
const STORAGE_KEY_LAST_GROUP_ID = 'uni_steps_last_group_id';

/**
 * 認証情報を提供するプロバイダーコンポーネントである．
 * URLパラメータおよび LocalStorage の両方からユーザーIDを復元・永続化する．
 */
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [searchParams] = useSearchParams();
  const userIdFromUrl = searchParams.get('user_id') || '';

  // 初期値: URLパラメータ優先，なければ localStorage から取得
  const [userId, setUserId] = useState<string>(() => {
    return userIdFromUrl || localStorage.getItem(STORAGE_KEY_USER_ID) || '';
  });

  // URL パラメータ変更時にレンダリング段階で state を同期（React 公式パターン）
  const [prevUrlUserId, setPrevUrlUserId] = useState(userIdFromUrl);
  if (userIdFromUrl !== prevUrlUserId) {
    setPrevUrlUserId(userIdFromUrl);
    if (userIdFromUrl) {
      setUserId(userIdFromUrl);
    }
  }

  // 外部システム（LocalStorage）への同期
  useEffect(() => {
    if (userId) {
      localStorage.setItem(STORAGE_KEY_USER_ID, userId);
    }
  }, [userId]);

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(() => !!userId);

  const fetchUser = useCallback(async (id: string, isRefresh: boolean = false) => {
    if (!id) {
      setUser(null);
      setLoading(false);
      return;
    }

    if (isRefresh) {
      setLoading(true);
    }

    const [userData, err] = await handle(userUsecase.fetchMe(id));
    if (err) {
      console.error("AuthContext: Failed to fetch user", err.message);
      localStorage.removeItem(STORAGE_KEY_USER_ID);
      localStorage.removeItem(STORAGE_KEY_LAST_GROUP_ID);
      setUser(null);
      setUserId('');
      setLoading(false);
      return;
    }

    setUser(userData);
    setLoading(false);
  }, []);

  useEffect(() => {
    if (!userId) return;
    // 非同期呼び出しで cascading render 警告を回避する
    Promise.resolve().then(() => fetchUser(userId, false));
  }, [userId, fetchUser]);

  const refreshUser = useCallback(async () => {
    await fetchUser(userId, true);
  }, [userId, fetchUser]);

  const logout = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY_USER_ID);
    localStorage.removeItem(STORAGE_KEY_LAST_GROUP_ID);
    setUser(null);
    setLoading(false);
    setUserId('');
  }, []);

  return (
    <AuthContext.Provider value={{ userId, user, loading, refreshUser, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
