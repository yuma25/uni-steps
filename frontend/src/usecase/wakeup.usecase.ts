import type { IWakeupPort } from '../port';
import type { WakeupCheck } from '../domain/models';

export class WakeupUsecase {
  readonly wakeupPort: IWakeupPort;

  constructor(wakeupPort: IWakeupPort) {
    this.wakeupPort = wakeupPort;
  }

  async createSchedule(params: { user_id: string; group_id: string; target_time: string; grace_minutes: number }): Promise<WakeupCheck> {
    return this.wakeupPort.request(params);
  }

  async checkIn(userId: string): Promise<void> {
    return this.wakeupPort.checkin(userId);
  }

  async cancelSchedule(userId: string): Promise<void> {
    return this.wakeupPort.cancel(userId);
  }

  async fetchUserActive(userId: string): Promise<WakeupCheck[]> {
    return this.wakeupPort.getActive(userId);
  }

  async fetchGroupActive(groupId: string): Promise<WakeupCheck[]> {
    return this.wakeupPort.getActiveByGroup(groupId);
  }
}
