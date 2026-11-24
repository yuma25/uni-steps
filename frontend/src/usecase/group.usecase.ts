import type { IGroupPort } from '../port';
import type { Group, NotificationLog } from '../domain/models';

export class GroupUsecase {
  readonly groupPort: IGroupPort;

  constructor(groupPort: IGroupPort) {
    this.groupPort = groupPort;
  }

  async createGroup(name: string, ownerId: string): Promise<Group> {
    return this.groupPort.createGroup(name, ownerId);
  }

  async joinGroup(inviteCode: string, userId: string): Promise<Group> {
    return this.groupPort.joinGroup(inviteCode, userId);
  }

  async fetchGroup(groupId: string): Promise<Group> {
    return this.groupPort.getGroup(groupId);
  }

  async fetchUserGroups(userId: string): Promise<Group[]> {
    return this.groupPort.listMyGroups(userId);
  }

  async updateSettings(
    groupId: string,
    name: string,
    intervals: number[],
    userId: string,
    aiCharacter: string,
    lineToken: string,
    lineGroupId: string,
    morningTime: string,
    eveningTime: string
  ): Promise<void> {
    return this.groupPort.updateSettings(
      groupId,
      name,
      intervals,
      userId,
      aiCharacter,
      lineToken,
      lineGroupId,
      morningTime,
      eveningTime
    );
  }

  async leaveGroup(groupId: string, userId: string): Promise<void> {
    return this.groupPort.leaveGroup(groupId, userId);
  }

  async transferOwnership(groupId: string, currentOwnerId: string, newOwnerId: string): Promise<void> {
    return this.groupPort.transferOwnership(groupId, currentOwnerId, newOwnerId);
  }

  async deleteGroup(groupId: string, userId: string): Promise<void> {
    return this.groupPort.deleteGroup(groupId, userId);
  }

  async getNotifications(groupId: string): Promise<NotificationLog[]> {
    return this.groupPort.listNotifications(groupId);
  }
}
