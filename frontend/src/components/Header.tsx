import React, { useState, useEffect } from 'react';
import { Bell, User, LogOut, Settings, RefreshCw } from 'lucide-react';
import { BankAccountService } from '../services/api';
import type { BankAccount } from '../types';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

interface HeaderProps {
  title: string;
  user?: User | null;
  onLogout?: () => void;
}

const Header: React.FC<HeaderProps> = ({ title, user, onLogout }) => {
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [account, setAccount] = useState<BankAccount | null>(null);
  const [loadingBalance, setLoadingBalance] = useState(false);

  useEffect(() => {
    if (user) {
      loadAccountBalance();
      // Auto-refresh balance every 3 seconds for real-time updates
      const interval = setInterval(loadAccountBalance, 3000);
      
      // Listen for balance update events from transactions
      const handleBalanceUpdate = () => {
        loadAccountBalance();
      };
      window.addEventListener('balanceUpdated', handleBalanceUpdate);
      
      return () => {
        clearInterval(interval);
        window.removeEventListener('balanceUpdated', handleBalanceUpdate);
      };
    }
  }, [user]);

  const loadAccountBalance = async () => {
    if (!user) return;
    
    try {
      setLoadingBalance(true);
      const response = await BankAccountService.getAccount(user.id);
      if (response.success && response.data) {
        setAccount(response.data);
      }
    } catch (error) {
      console.error('Failed to load account balance:', error);
    } finally {
      setLoadingBalance(false);
    }
  };

  return (
    <header className="bg-white border-b border-gray-200 px-6 py-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">{title}</h1>
        </div>
        
        <div className="flex items-center space-x-4">
          {/* Balance Display */}
          {user && account && (
            <div className="flex items-center space-x-2">
              <span className="text-xl font-bold text-gray-900">
                {account.balance.amount.toLocaleString('vi-VN')}
              </span>
              <span className="text-lg font-medium text-gray-600">
                {account.balance.currency}
              </span>
              {loadingBalance && (
                <RefreshCw className="h-4 w-4 text-gray-400 animate-spin" />
              )}
            </div>
          )}
          
          {/* Notifications */}
          <button className="p-2 text-gray-400 hover:text-gray-600 transition-colors">
            <Bell size={20} />
          </button>
          
          {/* User Profile */}
          {user && (
            <div className="relative">
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <div className="w-8 h-8 bg-indigo-600 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-medium">
                    {user.first_name.charAt(0)}{user.last_name.charAt(0)}
                  </span>
                </div>
                <div className="text-left">
                  <div className="text-sm font-medium text-gray-900">
                    {user.first_name} {user.last_name}
                  </div>
                  <div className="text-xs text-gray-500">{user.email}</div>
                </div>
              </button>

              {/* User Menu Dropdown */}
              {showUserMenu && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
                  <div className="px-4 py-2 border-b border-gray-200">
                    <p className="text-sm font-medium text-gray-900">
                      {user.first_name} {user.last_name}
                    </p>
                    <p className="text-xs text-gray-500">{user.email}</p>
                  </div>
                  
                  <button className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center space-x-2">
                    <Settings size={16} />
                    <span>Settings</span>
                  </button>
                  
                  {onLogout && (
                    <button
                      onClick={() => {
                        setShowUserMenu(false);
                        onLogout();
                      }}
                      className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center space-x-2"
                    >
                      <LogOut size={16} />
                      <span>Sign out</span>
                    </button>
                  )}
                </div>
              )}
            </div>
          )}

          {/* If no user, show simple profile icon */}
          {!user && (
            <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
              <User size={18} className="text-gray-600" />
            </div>
          )}
        </div>
      </div>

      {/* Click outside to close user menu */}
      {showUserMenu && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setShowUserMenu(false)}
        />
      )}
    </header>
  );
};

export default Header;