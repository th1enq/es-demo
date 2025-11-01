import { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import Dashboard from './components/Dashboard';
import AccountInfo from './components/AccountInfo';
import Payroll from './components/Payroll';
import Events from './components/Events';
import Login from './components/Login';
import Register from './components/Register';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
}

function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [currentView, setCurrentView] = useState<'login' | 'register' | 'dashboard'>('login');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in
    const token = localStorage.getItem('token');
    const savedUser = localStorage.getItem('user');
    
    if (token && savedUser) {
      try {
        const parsedUser = JSON.parse(savedUser);
        setUser(parsedUser);
        setIsAuthenticated(true);
        setCurrentView('dashboard');
      } catch (error) {
        // Invalid user data, clear storage
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      }
    }
    setIsLoading(false);
  }, []);

  const handleLoginSuccess = (_token: string, userData: User) => {
    setUser(userData);
    setIsAuthenticated(true);
    setCurrentView('dashboard');
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
    setIsAuthenticated(false);
    setCurrentView('login');
  };

  const handleRegisterSuccess = () => {
    setCurrentView('login');
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-600"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    if (currentView === 'register') {
      return (
        <Register
          onRegisterSuccess={handleRegisterSuccess}
          onSwitchToLogin={() => setCurrentView('login')}
        />
      );
    }
    return (
      <Login
        onLoginSuccess={handleLoginSuccess}
        onSwitchToRegister={() => setCurrentView('register')}
      />
    );
  }

  const getTabTitle = (tab: string) => {
    switch (tab) {
      case 'dashboard':
        return 'Dashboard';
      case 'accounts':
        return 'Account Information';
      case 'payroll':
        return 'Payroll Operations';
      case 'reports':
        return 'Reports';
      case 'events':
        return 'Events History';
      case 'contacts':
        return 'Contacts';
      default:
        return 'Dashboard';
    }
  };

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard':
        return <Dashboard />;
      case 'accounts':
        return <AccountInfo />;
      case 'payroll':
        return <Payroll />;
      case 'reports':
        return <div className="p-6"><h2 className="text-2xl font-semibold">Reports - Coming Soon</h2></div>;
      case 'events':
        return <Events />;
      case 'contacts':
        return <div className="p-6"><h2 className="text-2xl font-semibold">Contacts - Coming Soon</h2></div>;
      default:
        return <Dashboard />;
    }
  };

  return (
    <div className="h-screen flex bg-gray-50">
      <Sidebar activeTab={activeTab} onTabChange={setActiveTab} />
      <div className="flex-1 flex flex-col">
        <Header title={getTabTitle(activeTab)} user={user} onLogout={handleLogout} />
        <main className="flex-1 overflow-auto">
          {renderContent()}
        </main>
      </div>
    </div>
  );
}

export default App;