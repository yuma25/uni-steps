import { useState, useCallback } from 'react';
import type { Task, Group, WakeupCheck, NotificationLog } from '../../domain/models';
import { taskUsecase, groupUsecase, wakeupUsecase, userUsecase } from '../../di/container';
import { handle } from '../../utils/helpers';

/**
 * ダッシュボード画面のプレゼンター（状態管理・ViewModel・ユースケース呼び出し）
 */
export const useDashboardPresenter = (userId: string, groupId: string, initialGroup?: Group | null) => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [group, setGroup] = useState<Group | null>(() => {
    if (initialGroup) return initialGroup;
    try {
      const cached = localStorage.getItem(`cached_group_${groupId}`);
      if (cached) return JSON.parse(cached);
    } catch {
      // ignore
    }
    return null;
  });
  const [activeWakeup, setActiveWakeup] = useState<WakeupCheck | null>(null);
  const [groupWakeups, setGroupWakeups] = useState<WakeupCheck[]>([]);
  const [notificationLogs, setNotificationLogs] = useState<NotificationLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [serverTokenMissing, setServerTokenMissing] = useState<boolean>(false);
  const [isProcessing, setIsProcessing] = useState(false);

  // データの一括取得
  const fetchData = useCallback(async () => {
    if (!userId || !groupId) return;

    setError(null);
    const [results, err] = await handle(Promise.allSettled([
      taskUsecase.fetchGroupTasks(groupId),
      groupUsecase.fetchGroup(groupId),
      wakeupUsecase.fetchUserActive(userId),
      userUsecase.fetchMe(userId),
      wakeupUsecase.fetchGroupActive(groupId),
      groupUsecase.getNotifications(groupId)
    ]));

    if (err) {
      console.error("useDashboardPresenter: Fetch error:", err.message);
      setError("データの取得中に予期せぬエラーが発生しました。");
      setLoading(false);
      return;
    }

    const [taskDataRes, groupRes, wakeupsRes, userRes, groupWakeupsRes, logsRes] = results;

    if (taskDataRes.status === 'fulfilled') {
      setTasks(taskDataRes.value || []);
    }

    if (groupRes.status === 'fulfilled' && groupRes.value) {
      setGroup(groupRes.value);
      try {
        localStorage.setItem(`cached_group_${groupId}`, JSON.stringify(groupRes.value));
      } catch {
        // ignore
      }
    }

    if (wakeupsRes.status === 'fulfilled') {
      const pendingWakeup = wakeupsRes.value?.find(w => w.status === 'pending');
      setActiveWakeup(pendingWakeup || null);
    }

    if (userRes.status === 'fulfilled') {
      setServerTokenMissing(!userRes.value.has_push_token);
    }

    if (groupWakeupsRes.status === 'fulfilled') {
      setGroupWakeups(groupWakeupsRes.value || []);
    }

    if (logsRes.status === 'fulfilled') {
      setNotificationLogs(logsRes.value || []);
    }

    setLoading(false);
  }, [groupId, userId]);

  // タスク完了切り替え
  const toggleTaskCompletion = useCallback(async (taskId: string) => {
    const [, err] = await handle(taskUsecase.toggleTask(taskId, userId));
    if (!err) {
      await fetchData();
    }
    return !err;
  }, [fetchData, userId]);

  // タスク保存・更新
  const saveTask = useCallback(async (taskData: Partial<Task>, editingTaskId?: string) => {
    setIsProcessing(true);
    const [result, err] = editingTaskId
      ? await handle(taskUsecase.updateTask(editingTaskId, taskData, userId))
      : await handle(taskUsecase.createManualTask({ ...taskData, group_id: groupId }));

    setIsProcessing(false);
    if (!err && result) {
      await fetchData();
      return true;
    }
    return false;
  }, [fetchData, groupId, userId]);

  // タスク削除
  const deleteTask = useCallback(async (taskId: string) => {
    setIsProcessing(true);
    const [, err] = await handle(taskUsecase.deleteTask(taskId, userId));
    setIsProcessing(false);
    if (!err) {
      await fetchData();
      return true;
    }
    return false;
  }, [fetchData, userId]);

  // Classroom 同期
  const syncClassroomTasks = useCallback(async () => {
    setIsProcessing(true);
    const [result, err] = await handle(taskUsecase.syncTasks(userId, groupId));
    setIsProcessing(false);
    if (!err && result) {
      await fetchData();
      return { success: true, count: result.tasks.length };
    }
    return { success: false, count: 0 };
  }, [fetchData, groupId, userId]);

  // グループ設定更新
  const updateGroupSettings = useCallback(async (
    name: string,
    intervals: number[],
    aiCharacter: string,
    lineToken: string,
    lineGroupId: string,
    morningTime: string,
    eveningTime: string
  ) => {
    setIsProcessing(true);
    const [, err] = await handle(groupUsecase.updateSettings(
      groupId,
      name,
      intervals,
      userId,
      aiCharacter,
      lineToken,
      lineGroupId,
      morningTime,
      eveningTime
    ));
    setIsProcessing(false);
    if (!err) {
      await fetchData();
      return true;
    }
    return false;
  }, [fetchData, groupId, userId]);

  // オーナー権限譲渡
  const transferOwnership = useCallback(async (newOwnerId: string) => {
    setIsProcessing(true);
    const [, err] = await handle(groupUsecase.transferOwnership(groupId, userId, newOwnerId));
    setIsProcessing(false);
    return !err;
  }, [groupId, userId]);

  // グループ退出
  const leaveGroup = useCallback(async () => {
    setIsProcessing(true);
    const [, err] = await handle(groupUsecase.leaveGroup(groupId, userId));
    setIsProcessing(false);
    return !err;
  }, [groupId, userId]);

  // グループ削除
  const deleteGroup = useCallback(async () => {
    setIsProcessing(true);
    const [, err] = await handle(groupUsecase.deleteGroup(groupId, userId));
    setIsProcessing(false);
    return !err;
  }, [groupId, userId]);

  // 起床見守り予約
  const requestWakeup = useCallback(async (targetTime: string, graceMinutes: number) => {
    setIsProcessing(true);
    const [, err] = await handle(wakeupUsecase.createSchedule({
      user_id: userId,
      group_id: groupId,
      target_time: targetTime,
      grace_minutes: graceMinutes
    }));
    setIsProcessing(false);
    if (!err) {
      await fetchData();
      return true;
    }
    return false;
  }, [fetchData, groupId, userId]);

  // 起床チェックイン
  const checkinWakeup = useCallback(async () => {
    setIsProcessing(true);
    const [, err] = await handle(wakeupUsecase.checkIn(userId));
    setIsProcessing(false);
    if (!err) {
      await fetchData();
      return true;
    }
    return false;
  }, [fetchData, userId]);

  // 起床見守りキャンセル
  const cancelWakeup = useCallback(async () => {
    setIsProcessing(true);
    const [, err] = await handle(wakeupUsecase.cancelSchedule(userId));
    setIsProcessing(false);
    if (!err) {
      await fetchData();
      return true;
    }
    return false;
  }, [fetchData, userId]);

  return {
    tasks,
    group,
    activeWakeup,
    groupWakeups,
    notificationLogs,
    loading,
    error,
    serverTokenMissing,
    setServerTokenMissing,
    isProcessing,
    fetchData,
    toggleTaskCompletion,
    saveTask,
    deleteTask,
    syncClassroomTasks,
    updateGroupSettings,
    transferOwnership,
    leaveGroup,
    deleteGroup,
    requestWakeup,
    checkinWakeup,
    cancelWakeup,
  };
};

export { useDashboardPresenter as useDashboardData };
