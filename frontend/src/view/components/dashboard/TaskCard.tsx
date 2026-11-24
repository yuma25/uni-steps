import { GraduationCap, Globe } from 'lucide-react';
import type { Task } from '../../../domain/models';
import { formatDate } from '../../../utils/helpers';

interface TaskCardProps {
  task: Task;
  userId: string;
  onEdit: (task: Task) => void;
  onToggleCompletion: (taskId: string) => void;
}

const TaskCard = ({ task, userId, onEdit, onToggleCompletion }: TaskCardProps) => {
  const myProgress = task.user_progress?.find(p => p.user_id === userId);
  const deadlineDate = new Date(task.deadline);
  const isUndetermined = deadlineDate.getFullYear() <= 1;
  const isPast = !isUndetermined && deadlineDate < new Date();

  const getStatus = () => {
    if (myProgress?.is_completed) return { label: "完了", className: "status-completed" };
    if (isPast) {
      return myProgress 
        ? { label: "提出遅れ", className: "status-overdue" } 
        : { label: "終了", className: "status-closed" };
    }
    return { label: "未完了", className: "status-pending" };
  };

  const status = getStatus();
  const canEdit = !(task.source === 'manual' && isPast);
  const isLMS = task.source === 'google_classroom';

  const recurrenceLabel = task.recurrence?.type === 'weekly'
    ? '毎週'
    : task.recurrence?.type === 'biweekly'
    ? '隔週'
    : null;

  return (
    <div className="task-card">
      <div className="task-info">
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px', flexWrap: 'wrap', marginBottom: '6px' }}>
          {task.subject && (
            <span style={{
              fontSize: '0.75rem', fontWeight: 700,
              color: 'var(--brand, #2e7d32)', backgroundColor: '#ecfdf5',
              padding: '2px 8px', borderRadius: '6px'
            }}>
              {task.subject}
            </span>
          )}
          {recurrenceLabel && (
            <span style={{
              fontSize: '0.75rem', fontWeight: 700,
              color: '#4338ca', backgroundColor: '#e0e7ff',
              padding: '2px 8px', borderRadius: '6px'
            }}>
              {recurrenceLabel}
            </span>
          )}
        </div>
        <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start'}}>
          <h3>{task.title}</h3>
          {isLMS ? (
            <span title="Google Classroom"><GraduationCap size={18} color="var(--brand)" /></span>
          ) : (
            <span title="手動課題"><Globe size={18} color="var(--text-tertiary)" /></span>
          )}
        </div>
        <div className="deadline">
          <span>{formatDate(task.deadline)}</span>
        </div>
        
        <div style={{display: 'flex', flexWrap: 'wrap', gap: '6px', marginTop: '1.25rem'}}>
          {task.user_progress?.map(p => {
            const displayName = p.user_name || "メンバー";
            return (
              <div 
                key={p.user_id} 
                className={`member-badge ${p.is_completed ? 'completed' : ''}`}
                style={{cursor: 'default'}}
                title={displayName}
              >
                <div style={{
                  width: '18px', height: '18px', borderRadius: '50%', 
                  background: p.is_completed ? 'var(--success)' : 'var(--neutral-300)',
                  color: 'white', display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: '10px', fontWeight: 800, marginRight: '6px'
                }}>
                  {displayName.charAt(0)}
                </div>
                {displayName}
              </div>
            );
          })}
        </div>
      </div>
      
      <div className="task-footer">
        <span className={`status-badge ${status.className}`}>{status.label}</span>
        <div style={{display: 'flex', gap: '8px'}}>
          {canEdit && !(task.source === 'manual' && isPast) && (
            <button onClick={() => onEdit(task)} className="btn btn-ghost" style={{padding: '6px 12px', fontSize: '0.8rem'}}>
              編集
            </button>
          )}
          {task.source === 'manual' && !isPast && (
            <button 
              onClick={() => onToggleCompletion(task.id)} 
              className={`btn ${status.label === "完了" ? 'btn-ghost' : 'btn-primary'}`}
              style={{padding: '6px 16px', fontSize: '0.8rem', background: status.label === "完了" ? '' : 'var(--brand)'}}
            >
              {status.label === "完了" ? "戻す" : "完了"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default TaskCard;
