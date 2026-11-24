interface BottomNavProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const BottomNav = ({ activeTab, onTabChange }: BottomNavProps) => {
  return (
    <nav className="bottom-nav">
      <div className="sidebar-logo">
        <span>Uni-Steps</span>
      </div>

      <button 
        onClick={() => onTabChange('tasks')} 
        className={`nav-item ${activeTab === 'tasks' ? 'active' : ''}`}
      >
        <span>ホーム</span>
      </button>
      <button 
        onClick={() => onTabChange('wakeup')} 
        className={`nav-item ${activeTab === 'wakeup' ? 'active' : ''}`}
      >
        <span>起床見守り</span>
      </button>
      <button 
        onClick={() => onTabChange('settings')} 
        className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
      >
        <span>設定</span>
      </button>
    </nav>
  );
};

export default BottomNav;
