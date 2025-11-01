import React, { useState } from 'react';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import Dashboard from './components/Dashboard';
import AccountManagement from './components/AccountManagement';
import AccountOperations from './components/AccountOperations';

function App() {
  const [activeTab, setActiveTab] = useState('dashboard');

  const getTabTitle = (tab: string) => {
    switch (tab) {
      case 'dashboard':
        return 'Dashboard';
      case 'accounts':
        return 'Bank Accounts';
      case 'payroll':
        return 'Payroll';
      case 'reports':
        return 'Reports';
      case 'advisor':
        return 'Advisor';
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
        return <AccountManagement />;
      case 'payroll':
        return <div className="p-6"><h2 className="text-2xl font-semibold">Payroll - Coming Soon</h2></div>;
      case 'reports':
        return <div className="p-6"><h2 className="text-2xl font-semibold">Reports - Coming Soon</h2></div>;
      case 'advisor':
        return <div className="p-6"><h2 className="text-2xl font-semibold">Advisor - Coming Soon</h2></div>;
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
        <Header title={getTabTitle(activeTab)} />
        <main className="flex-1 overflow-auto">
          {renderContent()}
        </main>
      </div>
    </div>
  );
}

export default App;