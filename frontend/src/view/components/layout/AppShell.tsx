import type { ReactNode } from 'react';
import BottomNav from '../dashboard/BottomNav';

interface AppShellProps {
  children: ReactNode;
  activeTab: string;
  onTabChange: (tab: string) => void;
  header: ReactNode;
}

const AppShell = ({ children, activeTab, onTabChange, header }: AppShellProps) => {
  return (
    <div className="app-shell">
      <BottomNav activeTab={activeTab} onTabChange={onTabChange} />
      <main className="app-main">
        {header}
        <div style={{marginTop: '1rem'}}>
          {children}
        </div>
      </main>
    </div>
  );
};

export default AppShell;
