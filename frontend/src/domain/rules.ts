import type { Task } from './models';

/**
 * 課題の期限が切れているかを判定する純粋なドメインルール関数である．
 */
export function isOverdue(task: Task, now: Date = new Date()): boolean {
  if (!task.deadline) return false;
  const deadline = new Date(task.deadline);
  if (deadline.getFullYear() <= 1) return false;
  return deadline.getTime() < now.getTime();
}

/**
 * 課題が特定ユーザーにおいて完了しているかを判定するドメインルール関数である．
 */
export function isTaskCompletedByUser(task: Task, userId: string): boolean {
  if (!task.user_progress) return false;
  const progress = task.user_progress.find((p) => p.user_id === userId);
  return progress ? progress.is_completed : false;
}
