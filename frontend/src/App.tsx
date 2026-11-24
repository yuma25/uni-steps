import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './view/pages/LoginPage';
import DashboardPage from './view/pages/DashboardPage';
import GroupSelectionPage from './view/pages/GroupSelectionPage';
import { AuthProvider } from './view/context/AuthContext';
import './App.css';

// アプリケーションのメインコンポーネントである．
// ルーティングの設定を行う．
function App() {
  return (
    <Router>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<LoginPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/select-group" element={<GroupSelectionPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
        </Routes>
      </AuthProvider>
    </Router>
  );
}

export default App;
