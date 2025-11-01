import React, { useState } from 'react';
import { Activity, DollarSign, Minus, Plus } from 'lucide-react';
import toast, { Toaster } from 'react-hot-toast';
import { BankAccountService } from '../services/api';
import type { BankAccount, EventsHistoryResponse } from '../types';

const AccountOperations: React.FC = () => {
  const [accountId, setAccountId] = useState('');
  const [account, setAccount] = useState<BankAccount | null>(null);
  const [events, setEvents] = useState<EventsHistoryResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [depositAmount, setDepositAmount] = useState('');
  const [withdrawAmount, setWithdrawAmount] = useState('');

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

  const handleGetAccount = async () => {
    if (!accountId.trim()) {
      toast.error('Please enter an account ID');
      return;
    }

    setLoading(true);
    try {
      const response = await BankAccountService.getAccount(accountId);
      if (response.success) {
        setAccount(response.data || null);
        toast.success('Account loaded successfully');
      } else {
        toast.error(response.error?.message || 'Failed to get account');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error?.message || 'Failed to get account');
    } finally {
      setLoading(false);
    }
  };

  const handleDeposit = async () => {
    if (!accountId.trim() || !depositAmount) {
      toast.error('Please enter account ID and deposit amount');
      return;
    }

    setLoading(true);
    try {
      const paymentId = generatePaymentId();
      const response = await BankAccountService.deposit(accountId, { 
        amount: parseFloat(depositAmount),
        payment_id: paymentId
      });
      if (response.success) {
        toast.success(`Deposit successful (${parseFloat(depositAmount).toLocaleString('vi-VN')} VND - Payment ID: ${paymentId.slice(0, 8)}...)`);
        setDepositAmount('');
        // Refresh account data
        handleGetAccount();
      } else {
        toast.error(response.error?.message || 'Deposit failed');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error?.message || 'Deposit failed');
    } finally {
      setLoading(false);
    }
  };

  const handleWithdraw = async () => {
    if (!accountId.trim() || !withdrawAmount) {
      toast.error('Please enter account ID and withdraw amount');
      return;
    }

    setLoading(true);
    try {
      const paymentId = generatePaymentId();
      const response = await BankAccountService.withdraw(accountId, { 
        amount: parseFloat(withdrawAmount),
        payment_id: paymentId
      });
      if (response.success) {
        toast.success(`Withdrawal successful (${parseFloat(withdrawAmount).toLocaleString('vi-VN')} VND - Payment ID: ${paymentId.slice(0, 8)}...)`);
        setWithdrawAmount('');
        // Refresh account data
        handleGetAccount();
      } else {
        toast.error(response.error?.message || 'Withdrawal failed');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error?.message || 'Withdrawal failed');
    } finally {
      setLoading(false);
    }
  };

  const handleGetEvents = async () => {
    if (!accountId.trim()) {
      toast.error('Please enter an account ID');
      return;
    }

    setLoading(true);
    try {
      const response = await BankAccountService.getEventsHistory(accountId);
      if (response.success) {
        setEvents(response.data || null);
        toast.success('Events loaded successfully');
      } else {
        toast.error(response.error?.message || 'Failed to get events');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error?.message || 'Failed to get events');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <Toaster position="top-right" />
      
      {/* Account ID Input */}
      <div className="card p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Account Operations</h3>
        <div className="flex space-x-4 mb-4">
          <input
            type="text"
            placeholder="Enter Account ID (UUID)"
            value={accountId}
            onChange={(e) => setAccountId(e.target.value)}
            className="input-field flex-1"
          />
          <button
            onClick={handleGetAccount}
            disabled={loading}
            className="btn-primary"
          >
            Load Account
          </button>
        </div>
      </div>

      {/* Account Display */}
      {account && (
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Account Details</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-gray-500">Name</p>
              <p className="font-medium">{account.firstName} {account.lastName}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Email</p>
              <p className="font-medium">{account.email}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Balance</p>
              <p className="font-medium text-2xl text-green-600">
                {account.balance.amount.toLocaleString('vi-VN')} {account.balance.currency}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Operations */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Deposit */}
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <Plus className="text-green-600 mr-2" size={20} />
            Deposit Money
          </h3>
          <div className="space-y-4">
            <input
              type="number"
              placeholder="Enter amount (VND)"
              value={depositAmount}
              onChange={(e) => setDepositAmount(e.target.value)}
              className="input-field"
              min="0"
              step="1000"
            />
            <button
              onClick={handleDeposit}
              disabled={loading || !accountId}
              className="btn-primary w-full flex items-center justify-center space-x-2"
            >
              <DollarSign size={16} />
              <span>Deposit</span>
            </button>
          </div>
        </div>

        {/* Withdraw */}
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <Minus className="text-red-600 mr-2" size={20} />
            Withdraw Money
          </h3>
          <div className="space-y-4">
            <input
              type="number"
              placeholder="Enter amount (VND)"
              value={withdrawAmount}
              onChange={(e) => setWithdrawAmount(e.target.value)}
              className="input-field"
              min="0"
              step="1000"
            />
            <button
              onClick={handleWithdraw}
              disabled={loading || !accountId}
              className="btn-primary w-full flex items-center justify-center space-x-2 bg-red-600 hover:bg-red-700"
            >
              <DollarSign size={16} />
              <span>Withdraw</span>
            </button>
          </div>
        </div>
      </div>

      {/* Events History */}
      <div className="card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 flex items-center">
            <Activity className="text-blue-600 mr-2" size={20} />
            Events History
          </h3>
          <button
            onClick={handleGetEvents}
            disabled={loading || !accountId}
            className="btn-secondary"
          >
            Load Events
          </button>
        </div>
        
        {events ? (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg">
              <p className="text-sm text-gray-600">
                Total Events: <span className="font-medium">{events.total_events}</span>
              </p>
              <p className="text-sm text-gray-600">
                Account ID: <span className="font-medium">{events.aggregate_id}</span>
              </p>
            </div>
            
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="text-left py-2 px-4 font-medium text-gray-500">Event Type</th>
                    <th className="text-left py-2 px-4 font-medium text-gray-500">Version</th>
                    <th className="text-left py-2 px-4 font-medium text-gray-500">Timestamp</th>
                    <th className="text-left py-2 px-4 font-medium text-gray-500">Data</th>
                  </tr>
                </thead>
                <tbody>
                  {events.events.map((event, index) => (
                    <tr key={index} className="border-b border-gray-100">
                      <td className="py-2 px-4 text-sm">{event.event_type}</td>
                      <td className="py-2 px-4 text-sm">{event.version}</td>
                      <td className="py-2 px-4 text-sm">
                        {new Date(event.timestamp).toLocaleString()}
                      </td>
                      <td className="py-2 px-4 text-sm">
                        <pre className="text-xs bg-gray-100 p-2 rounded">
                          {JSON.stringify(event.data, null, 2)}
                        </pre>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            Load an account and click "Load Events" to see the event history
          </div>
        )}
      </div>
    </div>
  );
};

export default AccountOperations;