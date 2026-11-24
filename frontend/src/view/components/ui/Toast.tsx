import { useEffect } from 'react';
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-react';

export type ToastType = 'success' | 'error' | 'info';

interface ToastProps {
  message: string | null;
  type?: ToastType;
  onClose: () => void;
  duration?: number;
}

/**
 * 操作の完了や通知をユーザーに直感的に伝えるトースト通知コンポーネント
 */
export const Toast = ({ message, type = 'success', onClose, duration = 3000 }: ToastProps) => {
  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => {
      onClose();
    }, duration);
    return () => clearTimeout(timer);
  }, [message, duration, onClose]);

  if (!message) return null;

  const iconMap = {
    success: <CheckCircle2 size={18} className="toast-icon-success" />,
    error: <AlertCircle size={18} className="toast-icon-error" />,
    info: <Info size={18} className="toast-icon-info" />,
  };

  return (
    <div className="toast-container" role="status" aria-live="polite">
      <div className={`toast-item toast-${type} animate-slide-down`}>
        <div className="toast-icon">{iconMap[type]}</div>
        <span className="toast-message">{message}</span>
        <button type="button" onClick={onClose} className="toast-close" aria-label="閉じる">
          <X size={14} />
        </button>
      </div>
    </div>
  );
};

export default Toast;
