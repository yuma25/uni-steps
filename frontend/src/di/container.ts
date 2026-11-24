import { taskApi } from '../infra/http/tasks';
import { groupApi } from '../infra/http/groups';
import { wakeupApi } from '../infra/http/wakeup';
import { userApi } from '../infra/http/user';
import { notificationApi } from '../infra/http/notifications';

import { TaskUsecase } from '../usecase/task.usecase';
import { GroupUsecase } from '../usecase/group.usecase';
import { WakeupUsecase } from '../usecase/wakeup.usecase';
import { UserUsecase } from '../usecase/user.usecase';
import { NotificationUsecase } from '../usecase/notification.usecase';

// シングルトンのユースケースインスタンス群（依存注入）
export const taskUsecase = new TaskUsecase(taskApi);
export const groupUsecase = new GroupUsecase(groupApi);
export const wakeupUsecase = new WakeupUsecase(wakeupApi);
export const userUsecase = new UserUsecase(userApi);
export const notificationUsecase = new NotificationUsecase(notificationApi);
