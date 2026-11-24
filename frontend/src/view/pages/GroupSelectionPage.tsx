import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, UserPlus, LogOut, Hash, X, ArrowRight, Sparkles } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useGroupSelectionPresenter } from '../../presenter/hooks/useGroupSelectionPresenter';

/**
 * グループ（部屋）選択画面（ポータル）の純粋表示コンポーネントである．
 */
const GroupSelectionPage = () => {
  const navigate = useNavigate();
  const { userId, logout, loading: authLoading } = useAuth();
  const { groups, loading, fetchGroups, createGroup, joinGroup } = useGroupSelectionPresenter(userId);

  const [newGroupName, setNewGroupName] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);

  useEffect(() => {
    if (!authLoading && !userId) {
      navigate('/login');
      return;
    }
    if (userId) {
      Promise.resolve().then(() => fetchGroups());
    }
  }, [userId, authLoading, navigate, fetchGroups]);

  // モーダル表示時に背景スクロールを禁止
  useEffect(() => {
    if (showCreateModal || showJoinModal) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [showCreateModal, showJoinModal]);

  const handleCreateGroup = async (e: FormEvent) => {
    e.preventDefault();
    if (!newGroupName.trim()) return;
    const newGroup = await createGroup(newGroupName);
    if (newGroup) {
      setShowCreateModal(false);
      localStorage.setItem('uni_steps_last_group_id', newGroup.id);
      navigate(`/dashboard?group_id=${newGroup.id}`, { state: { initialGroup: newGroup } });
    }
  };

  const handleJoinGroup = async (e: FormEvent) => {
    e.preventDefault();
    if (!inviteCode.trim()) return;
    const joinedGroup = await joinGroup(inviteCode);
    if (joinedGroup) {
      setShowJoinModal(false);
      localStorage.setItem('uni_steps_last_group_id', joinedGroup.id);
      navigate(`/dashboard?group_id=${joinedGroup.id}`, { state: { initialGroup: joinedGroup } });
    }
  };

  const selectGroup = (groupId: string) => {
    const targetGroup = groups.find(g => g.id === groupId);
    localStorage.setItem('uni_steps_last_group_id', groupId);
    navigate(`/dashboard?group_id=${groupId}`, { state: { initialGroup: targetGroup } });
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="portal-container">
      <div className="portal-content">
        <header className="selection-hero">
          <h1>Uni-Steps</h1>
          <p>参加する部屋を選択するか、新しく作成してください</p>
        </header>

        <div className="action-cards-grid">
          <button onClick={() => setShowCreateModal(true)} className="portal-card">
            <div className="icon-box" style={{ background: 'var(--brand-soft)', color: 'var(--brand)' }}>
              <Plus size={32} strokeWidth={3} />
            </div>
            <h3 style={{ fontSize: '1.4rem', fontWeight: 900, margin: '0 0 0.5rem 0' }}>新しい部屋を作成</h3>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', margin: 0, lineHeight: 1.5 }}>
              共同管理を始める新しい部屋を作ります
            </p>
          </button>

          <button onClick={() => setShowJoinModal(true)} className="portal-card">
            <div className="icon-box" style={{ background: '#dcfce7', color: 'var(--success)' }}>
              <UserPlus size={32} strokeWidth={3} />
            </div>
            <h3 style={{ fontSize: '1.4rem', fontWeight: 900, margin: '0 0 0.5rem 0' }}>招待コードで参加</h3>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', margin: 0, lineHeight: 1.5 }}>
              招待コードを入力して部屋に参加します
            </p>
          </button>
        </div>

        <section className="group-list-section">
          <div className="section-header" style={{ marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 900, color: 'var(--text-primary)', margin: 0 }}>
              所属している部屋
            </h2>
            <span style={{ fontSize: '0.8rem', fontWeight: 700, background: 'var(--surface-sunken)', padding: '2px 8px', borderRadius: '12px', color: 'var(--text-secondary)' }}>
              {groups.length}
            </span>
          </div>

          {loading ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-tertiary)' }}>
              読み込み中...
            </div>
          ) : groups.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem', background: 'var(--surface)', borderRadius: '24px', border: '2px dashed var(--border)' }}>
              <div style={{ margin: '0 auto 1rem', width: '48px', height: '48px', borderRadius: '50%', background: 'var(--brand-soft)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--brand)' }}>
                <Sparkles size={24} />
              </div>
              <p style={{ color: 'var(--text-secondary)', fontWeight: 600, margin: 0 }}>所属している部屋がありません</p>
            </div>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '1.25rem' }}>
              {groups.map((g) => (
                <button key={g.id} onClick={() => selectGroup(g.id)} className="group-item-card">
                  <div className="group-item-icon">
                    <Hash size={24} strokeWidth={2.5} />
                  </div>
                  <div style={{ flex: 1, textAlign: 'left', minWidth: 0 }}>
                    <div className="group-item-name">{g.name}</div>
                    {g.owner_id === userId ? (
                      <span style={{ fontSize: '0.7rem', fontWeight: 900, color: 'var(--brand)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                        Owner
                      </span>
                    ) : (
                      <span style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)' }}>
                        Member
                      </span>
                    )}
                  </div>
                  <ArrowRight size={18} color="var(--text-tertiary)" />
                </button>
              ))}
            </div>
          )}
        </section>

        <footer className="portal-footer" style={{ marginTop: '3rem', textAlign: 'center' }}>
          <button onClick={handleLogout} className="btn btn-ghost" style={{ padding: '10px 24px', borderRadius: '16px', fontSize: '0.9rem' }}>
            <LogOut size={16} />
            <span>ログアウト</span>
          </button>
        </footer>
      </div>

      {/* 部屋作成モーダル */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content animate-pop" onClick={(e) => e.stopPropagation()}>
            <button onClick={() => setShowCreateModal(false)} className="btn-ghost" style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none' }}>
              <X size={20} />
            </button>
            <div className="modal-header-text" style={{ marginBottom: '2rem' }}>
              <h2 style={{ fontWeight: 900 }}>新しい部屋を作成</h2>
            </div>
            <form onSubmit={handleCreateGroup}>
              <div className="form-group" style={{ marginBottom: '1.5rem' }}>
                <label><Hash size={14} /> 部屋の名前</label>
                <input
                  type="text"
                  placeholder="例：ゼミ用，月曜2限"
                  value={newGroupName}
                  onChange={(e) => setNewGroupName(e.target.value)}
                  autoFocus
                  required
                />
              </div>
              <button type="submit" disabled={loading || !newGroupName.trim()} className="btn btn-primary" style={{ width: '100%', padding: '14px', borderRadius: '14px' }}>
                {loading ? '作成中...' : '作成して移動'}
              </button>
            </form>
          </div>
        </div>
      )}

      {/* 部屋参加モーダル */}
      {showJoinModal && (
        <div className="modal-overlay" onClick={() => setShowJoinModal(false)}>
          <div className="modal-content animate-pop" onClick={(e) => e.stopPropagation()}>
            <button onClick={() => setShowJoinModal(false)} className="btn-ghost" style={{ position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none' }}>
              <X size={20} />
            </button>
            <div className="modal-header-text" style={{ marginBottom: '2rem' }}>
              <h2 style={{ fontWeight: 900 }}>招待コードで参加</h2>
            </div>
            <form onSubmit={handleJoinGroup}>
              <div className="form-group" style={{ marginBottom: '1.5rem' }}>
                <label><UserPlus size={14} /> 招待コード</label>
                <input
                  type="text"
                  placeholder="例：a1b2c3d4"
                  value={inviteCode}
                  onChange={(e) => setInviteCode(e.target.value)}
                  autoFocus
                  required
                />
              </div>
              <button
                type="submit"
                disabled={loading || !inviteCode.trim()}
                className="btn btn-primary"
                style={{ width: '100%', padding: '14px', borderRadius: '14px', background: 'var(--success)' }}
              >
                {loading ? '参加処理中...' : '参加する'}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default GroupSelectionPage;
