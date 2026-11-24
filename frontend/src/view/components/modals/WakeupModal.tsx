import type { FormEvent, Dispatch, SetStateAction } from 'react';
import { X, Sunrise, Clock, Info } from 'lucide-react';

interface WakeupModalProps {
  show: boolean;
  onClose: () => void;
  formData: {
    target_time: string;
    grace_minutes: number;
  };
  setFormData: Dispatch<SetStateAction<{
    target_time: string;
    grace_minutes: number;
  }>>;
  onSave: (e: FormEvent) => void;
  loading: boolean;
}

const WakeupModal = ({
  show,
  onClose,
  formData,
  setFormData,
  onSave,
  loading
}: WakeupModalProps) => {
  if (!show) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-content wakeup-modal animate-pop">
        <button onClick={onClose} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
        <div style={{marginBottom: '2rem'}}>
          <h2 style={{margin: 0, fontWeight: 900, display: 'flex', alignItems: 'center', gap: '8px'}}>
            <Sunrise size={24} color="var(--brand)" /> 起床見守り設定
          </h2>
        </div>
        
        <form onSubmit={onSave}>
          <div className="form-group">
            <label><Clock size={14} /> 起床予定時刻</label>
            <div style={{display: 'flex', gap: '10px', alignItems: 'center'}}>
              <input 
                type="datetime-local" 
                required 
                value={formData.target_time} 
                onChange={e => setFormData({...formData, target_time: e.target.value})} 
                style={{flex: 1}}
              />
            </div>
          </div>
          
          <div className="form-group">
            <label><Info size={14} /> 猶予時間</label>
            <div style={{display: 'flex', gap: '10px', alignItems: 'center'}}>
              <input 
                type="number" 
                required 
                value={isNaN(formData.grace_minutes) ? '' : formData.grace_minutes} 
                onChange={e => {
                  const val = parseInt(e.target.value);
                  setFormData({...formData, grace_minutes: isNaN(val) ? 0 : val});
                }} 
                placeholder="15" 
                style={{flex: 1}}
              />
              <span style={{fontWeight: 700, color: 'var(--text-secondary)'}}>分</span>
            </div>
            <p style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.5rem'}}>
              ※ 猶予時間内にチェックインしない場合、グループへ通知されます
            </p>
          </div>
          
          <button 
            type="submit" 
            disabled={loading} 
            className="btn btn-primary" 
            style={{width: '100%', padding: '14px', borderRadius: '16px', background: 'var(--brand)', fontSize: '1rem', marginTop: '1.5rem'}}
          >
            {loading ? "予約中..." : "見守りを開始"}
          </button>
        </form>
      </div>
    </div>
  );
};

export default WakeupModal;
