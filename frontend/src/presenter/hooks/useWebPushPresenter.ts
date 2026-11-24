import { useState, useCallback } from 'react';
import { notificationUsecase } from '../../di/container';
import { urlBase64ToUint8Array, handle } from '../../utils/helpers';

/**
 * Web Push 通知の購読と状態管理を行うプレゼンターフックである．
 */
export const useWebPushPresenter = (userId: string, groupId: string) => {
  const [notifPermission, setNotifPermission] = useState<NotificationPermission>(
    typeof Notification !== 'undefined' ? Notification.permission : 'default'
  );

  const handleEnableNotifications = useCallback(async (onSuccess: () => void) => {
    if (typeof Notification === 'undefined') {
      alert("お使いのブラウザはプッシュ通知に対応していません．スマホの場合は「ホーム画面に追加」してからお試しください．");
      return;
    }
    const [permission, pErr] = await handle(Notification.requestPermission());
    if (pErr) {
      console.error("useWebPushPresenter: Permission request failed", pErr.message);
      alert("通知の設定に失敗しました．");
      return;
    }

    setNotifPermission(permission);
    
    if (permission === 'granted') {
      const [registration, rErr] = await handle(navigator.serviceWorker.getRegistration());
      if (rErr) {
        console.error("useWebPushPresenter: Get registration failed", rErr.message);
        return;
      }

      if (!registration) {
        const [, regErr] = await handle(navigator.serviceWorker.register('/sw.js'));
        if (regErr) {
          console.error("useWebPushPresenter: Registration failed", regErr.message);
          return;
        }
      }
      
      const [reg, readyErr] = await handle(navigator.serviceWorker.ready);
      if (readyErr) {
        console.error("useWebPushPresenter: Service worker not ready", readyErr.message);
        return;
      }
      
      // 既存の購読があれば解除
      const [existingSub, subErr] = await handle(reg.pushManager.getSubscription());
      if (!subErr && existingSub) {
        await handle(existingSub.unsubscribe());
      }

      // 公開鍵の変換
      const vapidPublicKey = 'BDj40a3LAnB-Tyemxggm-wYyuHbE_kadO6CqX6u-Ewyrkqi5ypr-txXJO7jflV_4VGa47paZU7DX_-0OPZy6Bx8';
      const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);

      const [subscription, sErr] = await handle(reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: convertedVapidKey as unknown as BufferSource
      }));

      if (sErr) {
        console.error("useWebPushPresenter: Subscription failed", sErr.message);
        alert("通知の購読に失敗しました．");
        return;
      }
      
      const [, apiErr] = await handle(notificationUsecase.subscribe(userId, subscription));
      if (apiErr) {
        console.error("useWebPushPresenter: API subscribe failed", apiErr.message);
        return;
      }

      onSuccess();
      alert("通知が有効になりました！AI からのリマインドが届くようになります．");
    }
  }, [userId]);

  const handleSilentResubscribe = useCallback(async (onSuccess: () => void) => {
    if (typeof Notification === 'undefined' || Notification.permission !== 'granted') return;

    const [reg, rErr] = await handle(navigator.serviceWorker.ready);
    if (rErr) return;
    
    const [existingSub, subErr] = await handle(reg.pushManager.getSubscription());
    if (!subErr && existingSub) {
      await handle(existingSub.unsubscribe());
    }

    const vapidPublicKey = 'BDj40a3LAnB-Tyemxggm-wYyuHbE_kadO6CqX6u-Ewyrkqi5ypr-txXJO7jflV_4VGa47paZU7DX_-0OPZy6Bx8';
    const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);

    const [subscription, sErr] = await handle(reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: convertedVapidKey as unknown as BufferSource
    }));

    if (sErr) return;
    const [, apiErr] = await handle(notificationUsecase.subscribe(userId, subscription));
    if (!apiErr) {
      onSuccess();
    }
  }, [userId]);

  const handleSendTestNotification = useCallback(async (aiCharacter: string) => {
    const [, err] = await handle(notificationUsecase.sendTestNotification(userId, aiCharacter, groupId));
    if (err) {
      alert("テスト通知の送信に失敗しました．");
    } else {
      alert("テスト通知を送信しました！数秒後に届くか確認してください．");
    }
  }, [groupId, userId]);

  return {
    notifPermission,
    handleEnableNotifications,
    handleSilentResubscribe,
    handleSendTestNotification,
  };
};

export { useWebPushPresenter as useWebPush };
