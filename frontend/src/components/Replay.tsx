import React, { useState, useEffect } from 'react';
import { ReplayService } from '../services/api';
import type { ElasticsearchAccount, ReplayResult, SystemSummary } from '../types';

interface ReplayProps {}

const Replay: React.FC<ReplayProps> = () => {
  const [accounts, setAccounts] = useState<ElasticsearchAccount[]>([]);
  const [summary, setSummary] = useState<SystemSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [replayLoading, setReplayLoading] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchFilters, setSearchFilters] = useState({
    firstName: '',
    lastName: '',
    email: '',
    status: ''
  });
  const [replayResult, setReplayResult] = useState<ReplayResult | null>(null);

  // Load accounts and summary on component mount
  useEffect(() => {
    loadAccountsAndSummary();
  }, []);

  const loadAccountsAndSummary = async () => {
    setLoading(true);
    setError(null);
    try {
      const [accountsResponse, summaryResponse] = await Promise.all([
        ReplayService.searchAccountsInElasticsearch(),
        ReplayService.getSystemSummary()
      ]);

      if (accountsResponse.success) {
        setAccounts(accountsResponse.data?.accounts || []);
      }

      if (summaryResponse.success) {
        setSummary(summaryResponse.data || null);
      }
    } catch (err) {
      console.error('Error loading data:', err);
      setError('Failed to load data from Elasticsearch');
    } finally {
      setLoading(false);
    }
  };

  const handleReplay = async (recreateIndex = false) => {
    setReplayLoading(true);
    setError(null);
    setReplayResult(null);
    
    try {
      const response = await ReplayService.replayAllEvents(recreateIndex);
      
      if (response.success) {
        setReplayResult(response.data || null);
        // Reload accounts and summary after replay
        await loadAccountsAndSummary();
      } else {
        setError(response.message || 'Failed to replay events');
      }
    } catch (err) {
      console.error('Error replaying events:', err);
      setError('Failed to replay events');
    } finally {
      setReplayLoading(false);
    }
  };

  const handleDeleteIndex = async () => {
    if (!window.confirm('Are you sure you want to delete all data from Elasticsearch? This action cannot be undone.')) {
      return;
    }

    setDeleteLoading(true);
    setError(null);
    
    try {
      const response = await ReplayService.deleteElasticsearchIndex();
      
      if (response.success) {
        setAccounts([]);
        setSummary(null);
        setReplayResult(null);
      } else {
        setError(response.message || 'Failed to delete index');
      }
    } catch (err) {
      console.error('Error deleting index:', err);
      setError('Failed to delete index');
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleSearch = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const filters = Object.fromEntries(
        Object.entries(searchFilters).filter(([_, value]) => value.trim() !== '')
      );
      
      const response = await ReplayService.searchAccountsInElasticsearch(filters);
      
      if (response.success) {
        setAccounts(response.data?.accounts || []);
      } else {
        setError(response.message || 'Failed to search accounts');
      }
    } catch (err) {
      console.error('Error searching accounts:', err);
      setError('Failed to search accounts');
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const formatDuration = (duration: number) => {
    return `${(duration / 1000000).toFixed(2)} ms`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">Event Sourcing Replay</h1>
        <div className="flex space-x-4">
          <button
            onClick={() => handleReplay(false)}
            disabled={replayLoading}
            className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white px-4 py-2 rounded-md transition-colors"
          >
            {replayLoading ? 'Replaying...' : 'Replay Events'}
          </button>
          <button
            onClick={() => handleReplay(true)}
            disabled={replayLoading}
            className="bg-green-600 hover:bg-green-700 disabled:bg-green-300 text-white px-4 py-2 rounded-md transition-colors"
          >
            {replayLoading ? 'Replaying...' : 'Replay with Index Recreation'}
          </button>
          <button
            onClick={handleDeleteIndex}
            disabled={deleteLoading}
            className="bg-red-600 hover:bg-red-700 disabled:bg-red-300 text-white px-4 py-2 rounded-md transition-colors"
          >
            {deleteLoading ? 'Deleting...' : 'Delete All Data'}
          </button>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
          {error}
        </div>
      )}

      {/* Replay Result */}
      {replayResult && (
        <div className="bg-green-50 border border-green-200 p-4 rounded-md">
          <h3 className="text-lg font-semibold text-green-800 mb-2">Replay Completed Successfully!</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="font-medium">Total Events:</span> {replayResult.totalEvents}
            </div>
            <div>
              <span className="font-medium">Processed:</span> {replayResult.processedEvents}
            </div>
            <div>
              <span className="font-medium">Accounts Created:</span> {replayResult.createdAccounts}
            </div>
            <div>
              <span className="font-medium">Duration:</span> {formatDuration(replayResult.duration)}
            </div>
          </div>
          {replayResult.errors && replayResult.errors.length > 0 && (
            <div className="mt-2">
              <span className="font-medium text-red-600">Errors:</span>
              <ul className="list-disc list-inside text-red-600 text-sm">
                {replayResult.errors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-sm font-medium text-gray-500">Total Accounts</h3>
            <p className="text-2xl font-bold text-gray-900">{summary.totalAccounts}</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-sm font-medium text-gray-500">Total Balance</h3>
            <p className="text-2xl font-bold text-green-600">{formatCurrency(summary.totalBalance, 'USD')}</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-sm font-medium text-gray-500">Total Transactions</h3>
            <p className="text-2xl font-bold text-blue-600">{summary.totalTransactions}</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-sm font-medium text-gray-500">Net Flow</h3>
            <p className="text-2xl font-bold text-purple-600">{formatCurrency(summary.netFlow, 'USD')}</p>
          </div>
        </div>
      )}

      {/* Search Filters */}
      <div className="bg-white p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4">Search Accounts</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
          <input
            type="text"
            placeholder="First Name"
            value={searchFilters.firstName}
            onChange={(e) => setSearchFilters({ ...searchFilters, firstName: e.target.value })}
            className="border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="text"
            placeholder="Last Name"
            value={searchFilters.lastName}
            onChange={(e) => setSearchFilters({ ...searchFilters, lastName: e.target.value })}
            className="border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="email"
            placeholder="Email"
            value={searchFilters.email}
            onChange={(e) => setSearchFilters({ ...searchFilters, email: e.target.value })}
            className="border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <select
            value={searchFilters.status}
            onChange={(e) => setSearchFilters({ ...searchFilters, status: e.target.value })}
            className="border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="frozen">Frozen</option>
          </select>
        </div>
        <div className="flex space-x-4">
          <button
            onClick={handleSearch}
            disabled={loading}
            className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white px-4 py-2 rounded-md transition-colors"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
          <button
            onClick={() => {
              setSearchFilters({ firstName: '', lastName: '', email: '', status: '' });
              loadAccountsAndSummary();
            }}
            className="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-md transition-colors"
          >
            Clear & Reload All
          </button>
        </div>
      </div>

      {/* Accounts Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold">Accounts from Elasticsearch ({accounts.length})</h3>
        </div>
        
        {loading ? (
          <div className="flex justify-center items-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : accounts.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No accounts found. Try replaying events to populate data.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Account
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Balance
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Transactions
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Total Deposits
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Total Withdrawals
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Last Activity
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {accounts.map((account) => (
                  <tr key={account.aggregateId} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">
                        {account.firstName} {account.lastName}
                      </div>
                      <div className="text-sm text-gray-500">
                        ID: {account.aggregateId.substring(0, 8)}...
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {account.email}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <span className={account.balance.amount >= 0 ? 'text-green-600' : 'text-red-600'}>
                        {formatCurrency(account.balance.amount, account.balance.currency)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {account.transactionCount}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600">
                      {formatCurrency(account.totalDeposits, account.balance.currency)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-red-600">
                      {formatCurrency(account.totalWithdrawals, account.balance.currency)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        account.status === 'active' 
                          ? 'bg-green-100 text-green-800'
                          : account.status === 'inactive'
                          ? 'bg-gray-100 text-gray-800'
                          : 'bg-red-100 text-red-800'
                      }`}>
                        {account.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {formatDate(account.lastActivity)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default Replay;