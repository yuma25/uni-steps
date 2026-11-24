import { ArrowLeft, RefreshCw, Plus } from 'lucide-react';
import type { Group } from '../../../domain/models';

interface DashboardHeaderProps {
  group: Group | null;
  onBack: () => void;
  onAddTask: () => void;
  onSync: () => void;
  loading: boolean;
  isSyncing?: boolean;
  lastSyncedAt?: string;
}

const DashboardHeader = ({
  group,
  onBack,
  onAddTask,
  onSync,
  loading,
  isSyncing = false,
  lastSyncedAt
}: DashboardHeaderProps) => {
  const copyInviteCode = () => {
    if (group?.invite_code) {
      navigator.clipboard.writeText(group.invite_code);
      alert("招待コードをコピーしました！");
    }
  };

  // メンバー名をカンマ区切りで生成する（最大 3 名まで表示，それ以上は「ほか X 名」）
  const memberNames = group?.users?.map(u => u.name) || [];
  const displayMembers = memberNames.length > 3 
    ? `${memberNames.slice(0, 3).join(', ')} ほか ${memberNames.length - 3} 名`
    : memberNames.join(', ');

  const formattedSyncTime = lastSyncedAt 
    ? new Date(lastSyncedAt).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })
    : null;

  return (
    <header className="dashboard-header">
      <div className="header-top-nav">
        <button onClick={onBack} className="back-button" title="部屋選択へ戻る">
          <ArrowLeft size={22} />
        </button>
        <div className="header-actions">
          {!loading && (
            <>
              <button 
                onClick={onSync} 
                className="btn btn-ghost" 
                disabled={isSyncing} 
                title={formattedSyncTime ? `最終同期: ${formattedSyncTime}` : "課題を同期"} 
                style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 14px' }}
              >
                <RefreshCw size={18} className={isSyncing ? "animate-spin" : ""} />
                <span className="hide-mobile" style={{ fontWeight: 700 }}>
                  {isSyncing ? "同期中" : "同期"}
                </span>
              </button>
              <button onClick={onAddTask} className="btn btn-primary" title="課題を追加">
                <Plus size={20} />
                <span className="hide-mobile">課題追加</span>
              </button>
            </>
          )}
        </div>
      </div>
      
      <div className="header-title-section">
        <h1>{group?.name || "Uni-Steps"}</h1>
        <div className="group-info" style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
          {group?.invite_code && (
            <button onClick={copyInviteCode} className="invite-code-pill" title="クリックしてコピー">
              コード: {group.invite_code}
            </button>
          )}
          {group?.users && (
            <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
              {displayMembers}
            </span>
          )}
          {formattedSyncTime && (
            <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-tertiary)' }}>
              最終同期 {formattedSyncTime}
            </span>
          )}
        </div>
      </div>
    </header>
  );
};

export default DashboardHeader;
