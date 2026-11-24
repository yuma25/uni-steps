import type { IUserPort } from '../port';
import type { User } from '../domain/models';

export class UserUsecase {
  readonly userPort: IUserPort;

  constructor(userPort: IUserPort) {
    this.userPort = userPort;
  }

  async fetchMe(userId: string): Promise<User> {
    return this.userPort.getMe(userId);
  }
}
