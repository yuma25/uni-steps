import type { ITaskPort } from '../port';
import type { Task } from '../domain/models';

export class TaskUsecase {
  readonly taskPort: ITaskPort;

  constructor(taskPort: ITaskPort) {
    this.taskPort = taskPort;
  }

  async fetchGroupTasks(groupId: string): Promise<Task[]> {
    return this.taskPort.listGroupTasks(groupId);
  }

  async toggleTask(taskId: string, userId: string): Promise<void> {
    return this.taskPort.toggleTaskCompletion(taskId, userId);
  }

  async createManualTask(task: Partial<Task>): Promise<Task> {
    return this.taskPort.createManualTask(task);
  }

  async updateTask(taskId: string, task: Partial<Task>, userId: string): Promise<Task | null> {
    return this.taskPort.updateTask(taskId, task, userId);
  }

  async deleteTask(taskId: string, userId: string): Promise<void> {
    return this.taskPort.deleteTask(taskId, userId);
  }

  async syncTasks(userId: string, groupId: string): Promise<{ message: string; tasks: Task[] }> {
    return this.taskPort.syncTasks(userId, groupId);
  }
}
