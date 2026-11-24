import type { INotificationPort } from '../port';

export class NotificationUsecase {
  readonly notificationPort: INotificationPort;

  constructor(notificationPort: INotificationPort) {
    this.notificationPort = notificationPort;
  }

  async subscribe(userId: string, subscription: PushSubscription): Promise<void> {
    return this.notificationPort.subscribe(userId, subscription);
  }

  async sendTestNotification(userId: string, aiCharacter: string, groupId?: string): Promise<void> {
    return this.notificationPort.sendTestNotification(userId, aiCharacter, groupId);
  }
}
