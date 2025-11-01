import React, { useState, useEffect } from 'react';
import { BankAccountService } from '../services/api';
import type { BankAccount } from '../types';
import { DollarSign, Plus, Minus, CreditCard, AlertCircle, CheckCircle } from 'lucide-react';

const Payroll: React.FC = () => {
  const [account, setAccount] = useState<BankAccount | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [depositAmount, setDepositAmount] = useState('');
  const [withdrawAmount, setWithdrawAmount] = useState('');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    loadAccountInfo();
  }, []);

  const loadAccountInfo = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const savedUser = localStorage.getItem('user');
      if (!savedUser) {
        setError('User information not found');
        return;
      }

      const user = JSON.parse(savedUser);
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

  const showMessage = (type: 'success' | 'error', text: string) => {
    setMessage({ type, text });
    setTimeout(() => setMessage(null), 5000);
  };

  const generatePaymentId = () => {
    // Tạo payment_id dạng UUID
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
      return crypto.randomUUID();
    }
    // Fallback UUID generation cho môi trường không hỗ trợ crypto.randomUUID
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  };

  const handleDeposit = async () => {
    if (!depositAmount || parseFloat(depositAmount) <= 0) {
      showMessage('error', 'Please enter a valid deposit amount');
      return;
    }

    if (!account) {
      showMessage('error', 'Account information not available');
      return;
    }

    setLoading(true);
    try {
      const paymentId = generatePaymentId();
      const response = await BankAccountService.deposit(account.aggregateID, {
        amount: parseFloat(depositAmount),
        payment_id: paymentId
      });

      if (response.success) {
        showMessage('success', `Successfully deposited ${parseFloat(depositAmount).toLocaleString('vi-VN')} VND (Payment ID: ${paymentId.slice(0, 8)}...)`);
        setDepositAmount('');
        // Reload account info to get updated balance
        await loadAccountInfo();
        // Trigger balance refresh in header
        window.dispatchEvent(new CustomEvent('balanceUpdated'));
      } else {
        showMessage('error', response.error?.message || 'Deposit failed');
      }
    } catch (err: any) {
      console.error('Error during deposit:', err);
      showMessage('error', err.response?.data?.error?.message || 'Deposit failed');
    } finally {
      setLoading(false);
    }
  };

  const handleWithdraw = async () => {
    if (!withdrawAmount || parseFloat(withdrawAmount) <= 0) {
      showMessage('error', 'Please enter a valid withdrawal amount');
      return;
    }

    if (!account) {
      showMessage('error', 'Account information not available');
      return;
    }

    const amount = parseFloat(withdrawAmount);
    if (amount > account.balance.amount) {
      showMessage('error', 'Insufficient balance for this withdrawal');
      return;
    }

    setLoading(true);
    try {
      const paymentId = generatePaymentId();
      const response = await BankAccountService.withdraw(account.aggregateID, {
        amount: amount,
        payment_id: paymentId
      });

      if (response.success) {
        showMessage('success', `Successfully withdrew ${parseFloat(withdrawAmount).toLocaleString('vi-VN')} VND (Payment ID: ${paymentId.slice(0, 8)}...)`);
        setWithdrawAmount('');
        // Reload account info to get updated balance
        await loadAccountInfo();
        // Trigger balance refresh in header
        window.dispatchEvent(new CustomEvent('balanceUpdated'));
      } else {
        showMessage('error', response.error?.message || 'Withdrawal failed');
      }
    } catch (err: any) {
      console.error('Error during withdrawal:', err);
      showMessage('error', err.response?.data?.error?.message || 'Withdrawal failed');
    } finally {
      setLoading(false);
    }
  };

  if (error && !account) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center space-x-2">
            <AlertCircle className="h-5 w-5 text-red-600" />
            <p className="text-red-800">{error}</p>
          </div>
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

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Payroll Operations</h2>
        <p className="text-gray-600">Manage deposits and withdrawals for your account</p>
      </div>

      {/* Message */}
      {message && (
        <div className={`mb-6 p-4 rounded-lg border ${
          message.type === 'success' 
            ? 'bg-green-50 border-green-200 text-green-800' 
            : 'bg-red-50 border-red-200 text-red-800'
        }`}>
          <div className="flex items-center space-x-2">
            {message.type === 'success' ? 
              <CheckCircle className="h-5 w-5" /> : 
              <AlertCircle className="h-5 w-5" />
            }
            <p>{message.text}</p>
          </div>
        </div>
      )}

      {/* Current Balance */}
      {account && (
        <div className="mb-6 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <CreditCard className="h-6 w-6 text-indigo-600" />
              <div>
                <p className="text-sm font-medium text-gray-500">Current Balance</p>
                <p className="text-2xl font-bold text-gray-900">
                  {account.balance.amount.toLocaleString('vi-VN')} {account.balance.currency}
                </p>
              </div>
            </div>
            <div className="text-right">
              <p className="text-sm text-gray-500">{account.firstName} {account.lastName}</p>
              <p className="text-sm text-gray-400">{account.email}</p>
            </div>
          </div>
        </div>
      )}

      {/* Operations */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Deposit */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center space-x-3 mb-4">
            <div className="p-2 bg-green-100 rounded-lg">
              <Plus className="h-6 w-6 text-green-600" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900">Deposit Money</h3>
              <p className="text-sm text-gray-500">Add funds to your account</p>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Amount (VND)
              </label>
              <div className="relative">
                <DollarSign className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                <input
                  type="number"
                  placeholder="0"
                  value={depositAmount}
                  onChange={(e) => setDepositAmount(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                  min="0"
                  step="1000"
                  disabled={loading}
                />
              </div>
            </div>

            <button
              onClick={handleDeposit}
              disabled={loading || !depositAmount || !account}
              className="w-full bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors flex items-center justify-center space-x-2"
            >
              {loading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              ) : (
                <>
                  <Plus className="h-4 w-4" />
                  <span>Deposit</span>
                </>
              )}
            </button>
          </div>
        </div>

        {/* Withdraw */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center space-x-3 mb-4">
            <div className="p-2 bg-red-100 rounded-lg">
              <Minus className="h-6 w-6 text-red-600" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900">Withdraw Money</h3>
              <p className="text-sm text-gray-500">Remove funds from your account</p>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Amount (VND)
              </label>
              <div className="relative">
                <DollarSign className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                <input
                  type="number"
                  placeholder="0"
                  value={withdrawAmount}
                  onChange={(e) => setWithdrawAmount(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                  min="0"
                  step="1000"
                  max={account?.balance.amount || 0}
                  disabled={loading}
                />
              </div>
              {account && (
                <p className="text-xs text-gray-500 mt-1">
                  Maximum: {account.balance.amount.toLocaleString('vi-VN')} VND
                </p>
              )}
            </div>

            <button
              onClick={handleWithdraw}
              disabled={loading || !withdrawAmount || !account}
              className="w-full bg-red-600 text-white py-2 px-4 rounded-lg hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors flex items-center justify-center space-x-2"
            >
              {loading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              ) : (
                <>
                  <Minus className="h-4 w-4" />
                  <span>Withdraw</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="mt-6 bg-gray-50 rounded-lg p-4">
        <h4 className="text-sm font-medium text-gray-700 mb-3">Quick Actions</h4>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
          {[100000, 500000, 1000000, 5000000].map(amount => (
            <button
              key={amount}
              onClick={() => setDepositAmount(amount.toString())}
              className="px-3 py-2 text-sm bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
              disabled={loading}
            >
              +{amount.toLocaleString('vi-VN')} VND
            </button>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Payroll;