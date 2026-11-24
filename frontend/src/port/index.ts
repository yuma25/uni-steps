import type { Task, Group, User, WakeupCheck, NotificationLog } from '../domain/models';

export interface ITaskPort {
  listGroupTasks(groupId: string): Promise<Task[]>;
  createManualTask(task: Partial<Task>): Promise<Task>;
  syncTasks(userId: string, groupId: string): Promise<{ message: string; tasks: Task[] }>;
  toggleTaskCompletion(taskId: string, userId: string): Promise<void>;
  updateTask(taskId: string, task: Partial<Task>, userId: string): Promise<Task | null>;
  deleteTask(taskId: string, userId: string): Promise<void>;
}

export interface IGroupPort {
  listMyGroups(userId: string): Promise<Group[]>;
  createGroup(name: string, ownerId: string): Promise<Group>;
  joinGroup(inviteCode: string, userId: string): Promise<Group>;
  updateSettings(
    groupId: string,
    name: string,
    intervals: number[],
    userId: string,
    aiCharacter: string,
    lineToken: string,
    lineGroupId: string,
    morningTime: string,
    eveningTime: string
  ): Promise<void>;
  leaveGroup(groupId: string, userId: string): Promise<void>;
  transferOwnership(groupId: string, currentOwnerId: string, newOwnerId: string): Promise<void>;
  deleteGroup(groupId: string, userId: string): Promise<void>;
  listNotifications(groupId: string): Promise<NotificationLog[]>;
}

export interface IWakeupPort {
  request(params: { user_id: string; group_id: string; target_time: string; grace_minutes: number }): Promise<WakeupCheck>;
  checkin(userId: string): Promise<void>;
  getActive(userId: string): Promise<WakeupCheck[]>;
  getActiveByGroup(groupId: string): Promise<WakeupCheck[]>;
  cancel(userId: string): Promise<void>;
}

export interface IUserPort {
  getMe(userId: string): Promise<User>;
}

export interface INotificationPort {
  subscribe(userId: string, subscription: PushSubscription): Promise<void>;
  sendTestNotification(userId: string, aiCharacter: string, groupId?: string): Promise<void>;
}
