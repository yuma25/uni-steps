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
  const userIdFromUrl = searchParams.get('user_id');

  // 初期値: URLパラメータ優先，なければ localStorage から取得
  const [userId, setUserId] = useState<string>(() => {
    if (userIdFromUrl) {
      localStorage.setItem(STORAGE_KEY_USER_ID, userIdFromUrl);
      return userIdFromUrl;
    }
    return localStorage.getItem(STORAGE_KEY_USER_ID) || '';
  });

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // URLパラメータに user_id が新しく付与された場合の同期
  useEffect(() => {
    if (userIdFromUrl && userIdFromUrl !== userId) {
      localStorage.setItem(STORAGE_KEY_USER_ID, userIdFromUrl);
      setUserId(userIdFromUrl);
    }
  }, [userIdFromUrl, userId]);

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
      // ユーザーが存在しない（無効なID）場合はストレージからも削除する
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
    Promise.resolve().then(() => fetchUser(userId, false));
  }, [userId, fetchUser]);

  const refreshUser = useCallback(async () => {
    await fetchUser(userId, true);
  }, [userId, fetchUser]);

  const logout = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY_USER_ID);
    localStorage.removeItem(STORAGE_KEY_LAST_GROUP_ID);
    setUser(null);
    setUserId('');
  }, []);

  return (
    <AuthContext.Provider value={{ userId, user, loading, refreshUser, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
