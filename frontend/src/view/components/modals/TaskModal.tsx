import { useState, type FormEvent, type Dispatch, type SetStateAction } from 'react';
import { X, CheckCircle2, AlertCircle, RefreshCw } from 'lucide-react';
import type { Task, Group } from '../../../domain/models';

interface TaskModalProps {
  show: boolean;
  onClose: () => void;
  editingTask: Task | null;
  formData: {
    title: string;
    subject: string;
    deadline: string;
    recurrence_type: string;
    assignees: string[];
  };
  setFormData: Dispatch<SetStateAction<{
    title: string;
    subject: string;
    deadline: string;
    recurrence_type: string;
    assignees: string[];
  }>>;
  onSave: (e: FormEvent) => void;
  onDelete: (taskId: string) => void;
  group: Group | null;
  loading: boolean;
  operatorId: string;
  ownerId: string;
}

const TaskModal = ({
  show,
  onClose,
  editingTask,
  formData,
  setFormData,
  onSave,
  onDelete,
  group,
  loading,
  operatorId,
  ownerId
}: TaskModalProps) => {
  const [actionType, setActionType] = useState<'save' | 'delete' | null>(null);

  if (!show) return null;

  // 権限チェック：作成者 または グループオーナーのみがタイトル・期限・削除の操作が可能
  const isAuthorized = !editingTask || editingTask.creator_id === operatorId || ownerId === operatorId;
  const isLMS = editingTask?.source === 'google_classroom';
  
  // 期限の編集制限：LMS課題かつ元々期限が設定されている場合は編集不可とする
  const isLMSDeadlineLocked = isLMS && editingTask?.is_lms_deadline_set;

  // 授業名（科目名）の編集制限：LMS課題で元々授業名が設定されている場合は編集不可とするが、未設定の場合は設定可能
  const isLMSSubjectLocked = isLMS && !!editingTask?.subject?.trim();

  // 担当者が一人も選択されていない場合は保存できないようにする．
  const isAssigneeEmpty = formData.assignees.length === 0;

  const handleSubmit = (e: FormEvent) => {
    setActionType('save');
    onSave(e);
  };

  const handleDelete = (taskId: string) => {
    setActionType('delete');
    onDelete(taskId);
  };

  const isSaving = loading && actionType === 'save';
  const isDeleting = loading && actionType === 'delete';

  return (
    <div className="modal-overlay">
      <div className="modal-content animate-pop">
        <button onClick={onClose} className="btn-ghost" style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none' }}>
          <X size={20} />
        </button>
        <div className="modal-header-text" style={{ marginBottom: '2rem' }}>
          <h2 style={{ margin: 0, fontWeight: 900 }}>{editingTask ? "課題を編集" : "課題の追加"}</h2>
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>タイトル</label>
            <input 
              type="text" 
              required 
              placeholder="例：レポート提出"
              value={formData.title} 
              onChange={e => setFormData({ ...formData, title: e.target.value })} 
              disabled={isLMS || !isAuthorized} 
            />
          </div>

          <div className="form-group">
            <label>授業名（任意）</label>
            <input 
              type="text" 
              placeholder="例：情報ネットワーク"
              value={formData.subject} 
              onChange={e => setFormData({ ...formData, subject: e.target.value })} 
              disabled={isLMSSubjectLocked || (!isLMS && !isAuthorized)} 
            />
            {isLMS && !isLMSSubjectLocked && (
              <p style={{ fontSize: '0.75rem', color: 'var(--brand)', marginTop: '0.4rem' }}>
                ※Google Classroom で授業名が未設定の課題です．Uni-Steps 上で授業名を設定・保存できます．
              </p>
            )}
            {isLMSSubjectLocked && (
              <p style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.4rem' }}>
                ※Google Classroom 側で設定された授業名です．
              </p>
            )}
          </div>
          
          <div className="form-group">
            <label>期限</label>
            <input 
              type="datetime-local" 
              value={formData.deadline} 
              onChange={e => setFormData({ ...formData, deadline: e.target.value })} 
              disabled={isLMSDeadlineLocked || !isAuthorized}
            />
            {!isAuthorized && <p style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.4rem' }}>※期限の変更は作成者またはオーナーのみ可能です．</p>}
            {isLMS && !isLMSDeadlineLocked && isAuthorized && <p style={{ fontSize: '0.75rem', color: 'var(--brand)', marginTop: '0.4rem' }}>※Google Classroom で期限未定の課題です．Uni-Steps 上で期限を設定・保存できます．</p>}
            {isLMSDeadlineLocked && isAuthorized && <p style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.4rem' }}>※Google Classroom 側で期限が設定されているため，ここでは変更できません．</p>}
          </div>

          <div className="form-group">
            <label>繰り返し</label>
            <div style={{ display: 'flex', gap: '8px', marginTop: '0.4rem' }}>
              {[
                { value: 'none', label: 'なし' },
                { value: 'weekly', label: '毎週' },
                { value: 'biweekly', label: '隔週' }
              ].map(opt => (
                <button
                  key={opt.value}
                  type="button"
                  disabled={!isAuthorized}
                  onClick={() => setFormData({ ...formData, recurrence_type: opt.value })}
                  style={{
                    flex: 1,
                    padding: '10px 14px',
                    borderRadius: '12px',
                    fontSize: '0.85rem',
                    fontWeight: 700,
                    border: formData.recurrence_type === opt.value ? '2px solid var(--brand)' : '1px solid var(--border)',
                    background: formData.recurrence_type === opt.value ? 'var(--brand-soft)' : 'white',
                    color: formData.recurrence_type === opt.value ? 'var(--brand)' : 'var(--text-secondary)',
                    cursor: !isAuthorized ? 'not-allowed' : 'pointer',
                    transition: 'all 0.2s'
                  }}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
          
          <div className="form-group">
            <label>担当者を編集</label>
            <div style={{
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fill, minmax(min(140px, 45%), 1fr))', 
              gap: '0.5rem', 
              marginTop: '0.5rem'
            }}>
              {group?.users?.map(member => {
                const isMe = member.id === operatorId;
                const canToggle = isAuthorized || isMe;
                
                return (
                  <button
                    key={member.id}
                    type="button"
                    onClick={() => {
                      if (isLMS || !canToggle) return;
                      const isSelected = formData.assignees.includes(member.id);
                      setFormData({ 
                        ...formData, 
                        assignees: isSelected 
                          ? formData.assignees.filter(id => id !== member.id) 
                          : [...formData.assignees, member.id] 
                      });
                    }}
                    style={{
                      display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 12px',
                      background: formData.assignees.includes(member.id) ? 'var(--brand-soft)' : '#f8fafc',
                      border: `2px solid ${formData.assignees.includes(member.id) ? 'var(--brand-light)' : '#f1f5f9'}`,
                      borderRadius: '12px', 
                      cursor: (isLMS || !canToggle) ? 'not-allowed' : 'pointer',
                      opacity: canToggle ? 1 : 0.6,
                      transition: 'all 0.2s', textAlign: 'left',
                      minWidth: 0
                    }}
                  >
                    <div style={{
                      width: '20px', height: '20px', borderRadius: '50%', flexShrink: 0,
                      background: formData.assignees.includes(member.id) ? 'var(--brand)' : 'var(--text-tertiary)',
                      display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontSize: '9px', fontWeight: 900
                    }}>
                      {formData.assignees.includes(member.id) ? <CheckCircle2 size={12} /> : member.name.charAt(0)}
                    </div>
                    <span style={{
                      fontWeight: 700, fontSize: '0.85rem', 
                      color: formData.assignees.includes(member.id) ? 'var(--brand)' : 'var(--text-secondary)',
                      whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis'
                    }}>{member.name}</span>
                  </button>
                );
              })}
            </div>
            {!isAuthorized && <p style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.8rem' }}>※作成者・オーナー以外は，自分自身の参加・辞退のみ操作可能です．</p>}
            {isAssigneeEmpty && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginTop: '0.75rem', color: 'var(--warning)', fontSize: '0.85rem', fontWeight: 600 }}>
                <AlertCircle size={14} />
                <span>担当者がいないため，保存するとこの課題は削除されます．</span>
              </div>
            )}
          </div>

          <div style={{ marginTop: '2.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <button 
              type="submit" 
              disabled={loading} 
              className="btn btn-primary" 
              style={{ width: '100%', padding: '16px', marginTop: '1.5rem', borderRadius: '16px', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}
            >
              {isSaving && <RefreshCw size={18} className="animate-spin" />}
              <span>{isSaving ? "保存中..." : (editingTask ? "更新を保存する" : "課題を登録する")}</span>
            </button>

            {editingTask && isAuthorized && !isLMS && (
              <button
                type="button"
                onClick={() => handleDelete(editingTask.id)}
                disabled={loading}
                className="btn btn-ghost"
                style={{ width: '100%', padding: '14px', borderRadius: '16px', color: 'var(--error)', borderColor: '#fee2e2', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}
              >
                {isDeleting && <RefreshCw size={18} className="animate-spin" />}
                <span>{isDeleting ? "削除中..." : "課題を削除する"}</span>
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;
