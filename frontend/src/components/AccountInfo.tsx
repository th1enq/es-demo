import React, { useState, useEffect } from 'react';
import { BankAccountService } from '../services/api';
import type { BankAccount } from '../types';
import { User, DollarSign, Mail, CreditCard, RefreshCw } from 'lucide-react';

const AccountInfo: React.FC = () => {
  const [account, setAccount] = useState<BankAccount | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAccountInfo();
  }, []);

  const loadAccountInfo = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Lấy user info từ localStorage
      const savedUser = localStorage.getItem('user');
      if (!savedUser) {
        setError('User information not found');
        return;
      }

      const user = JSON.parse(savedUser);
      // Sử dụng user ID làm bank account ID
      const response = await BankAccountService.getAccount(user.id);
      
      if (response.success && response.data) {
        setAccount(response.data);
      } else {
        setError(response.error?.message || 'Failed to load account information');
      }
    } catch (err: any) {
      console.error('Error loading account:', err);
      setError(err.response?.data?.error?.message || 'Failed to load account information');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-800">{error}</p>
          <button
            onClick={loadAccountInfo}
            className="mt-3 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!account) {
    return (
      <div className="p-6">
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <p className="text-yellow-800">No account information found</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Account Information</h2>
          <p className="text-gray-600">View your account details and current balance</p>
        </div>
        <button
          onClick={loadAccountInfo}
          disabled={loading}
          className="flex items-center space-x-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
        >
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        {/* Header */}
        <div className="bg-gradient-to-r from-indigo-600 to-purple-600 px-6 py-4">
          <div className="flex items-center space-x-3">
            <CreditCard className="h-8 w-8 text-white" />
            <div>
              <h3 className="text-xl font-semibold text-white">Bank Account</h3>
              <p className="text-indigo-100 text-sm">Account ID: {account.aggregateID}</p>
            </div>
          </div>
        </div>

        {/* Account Details */}
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Personal Information */}
            <div className="space-y-4">
              <div className="flex items-center space-x-3 p-4 bg-gray-50 rounded-lg">
                <Mail className="h-5 w-5 text-gray-500" />
                <div>
                  <p className="text-sm font-medium text-gray-500">Email</p>
                  <p className="text-lg text-gray-900">{account.email}</p>
                </div>
              </div>

              <div className="flex items-center space-x-3 p-4 bg-gray-50 rounded-lg">
                <User className="h-5 w-5 text-gray-500" />
                <div>
                  <p className="text-sm font-medium text-gray-500">Full Name</p>
                  <p className="text-lg text-gray-900">{account.firstName} {account.lastName}</p>
                </div>
              </div>
            </div>

            {/* Balance Information */}
            <div className="space-y-4">
              <div className="bg-gradient-to-br from-green-50 to-emerald-50 border border-green-200 rounded-lg p-6">
                <div className="flex items-center space-x-3 mb-2">
                  <DollarSign className="h-6 w-6 text-green-600" />
                  <p className="text-sm font-medium text-green-700">Current Balance</p>
                </div>
                <div className="flex items-baseline space-x-2">
                  <span className="text-3xl font-bold text-green-900">
                    {account.balance.amount.toLocaleString('vi-VN')}
                  </span>
                  <span className="text-lg font-medium text-green-700">
                    {account.balance.currency}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Account Summary */}
        <div className="bg-gray-50 px-6 py-4 border-t border-gray-200">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-gray-900">{account.balance.currency}</p>
              <p className="text-sm text-gray-500">Currency</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">Active</p>
              <p className="text-sm text-gray-500">Status</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">Standard</p>
              <p className="text-sm text-gray-500">Account Type</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">
                {new Date().getFullYear()}
              </p>
              <p className="text-sm text-gray-500">Member Since</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AccountInfo;