import React, { useState, useEffect } from 'react';
import { Plus, Search, Eye } from 'lucide-react';
import toast, { Toaster } from 'react-hot-toast';
import { BankAccountService } from '../services/api';
import type { BankAccount, CreateBankAccountRequest } from '../types';

const AccountManagement: React.FC = () => {
  const [_accounts, _setAccounts] = useState<BankAccount[]>([]);
  const [loading, setLoading] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // Form state
  const [formData, setFormData] = useState<CreateBankAccountRequest>({
    email: '',
    first_name: '',
    last_name: '',
    balance: 0,
    password: '',
  });

  const loadAccounts = async () => {
    setLoading(true);
    try {
      // Since we don't have a list endpoint, we'll show the create form
      // In a real application, you would have a GET /api/v1/bank_accounts endpoint
      setLoading(false);
    } catch (error) {
      toast.error('Failed to load accounts');
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAccounts();
  }, []);

  const handleCreateAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await BankAccountService.createAccount(formData);
      if (response.success) {
        toast.success('Bank account created successfully!');
        setShowCreateForm(false);
        setFormData({
          email: '',
          first_name: '',
          last_name: '',
          balance: 0,
          password: '',
        });
        loadAccounts();
      } else {
        toast.error(response.error?.message || 'Failed to create account');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error?.message || 'Failed to create account');
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name === 'balance' ? parseFloat(value) || 0 : value,
    }));
  };

  return (
    <div className="p-6 space-y-6">
      <Toaster position="top-right" />
      
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold text-gray-900">Bank Accounts</h2>
        <button
          onClick={() => setShowCreateForm(true)}
          className="btn-primary flex items-center space-x-2"
        >
          <Plus size={20} />
          <span>Create Account</span>
        </button>
      </div>

      {/* Search Bar */}
      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
        <input
          type="text"
          placeholder="Search accounts..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="input-field pl-10"
        />
      </div>

      {/* Create Account Modal */}
      {showCreateForm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Create New Bank Account</h3>
            
            <form onSubmit={handleCreateAccount} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                <input
                  type="email"
                  name="email"
                  value={formData.email}
                  onChange={handleInputChange}
                  className="input-field"
                  required
                />
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">First Name</label>
                  <input
                    type="text"
                    name="first_name"
                    value={formData.first_name}
                    onChange={handleInputChange}
                    className="input-field"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
                  <input
                    type="text"
                    name="last_name"
                    value={formData.last_name}
                    onChange={handleInputChange}
                    className="input-field"
                    required
                  />
                </div>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Initial Balance</label>
                <input
                  type="number"
                  name="balance"
                  value={formData.balance}
                  onChange={handleInputChange}
                  className="input-field"
                  min="0"
                  step="0.01"
                  required
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
                <input
                  type="password"
                  name="password"
                  value={formData.password}
                  onChange={handleInputChange}
                  className="input-field"
                  minLength={6}
                  required
                />
              </div>
              
              <div className="flex space-x-4 pt-4">
                <button
                  type="submit"
                  disabled={loading}
                  className="btn-primary flex-1"
                >
                  {loading ? 'Creating...' : 'Create Account'}
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateForm(false)}
                  className="btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Demo Account Operations */}
      <div className="card p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Account Operations Demo</h3>
        <div className="space-y-4">
          <p className="text-gray-600">
            To test account operations, create an account first, then use the account ID to perform operations.
          </p>
          
          <div className="bg-gray-50 p-4 rounded-lg">
            <h4 className="font-medium text-gray-900 mb-2">Available Operations:</h4>
            <ul className="text-sm text-gray-600 space-y-1">
              <li>• <strong>Create Account:</strong> POST /api/v1/bank_accounts</li>
              <li>• <strong>Get Account:</strong> GET /api/v1/bank_accounts/{'{id}'}</li>
              <li>• <strong>Deposit:</strong> POST /api/v1/bank_accounts/{'{id}'}/deposite</li>
              <li>• <strong>Withdraw:</strong> POST /api/v1/bank_accounts/{'{id}'}/withdraw</li>
              <li>• <strong>Events History:</strong> GET /api/v1/bank_accounts/{'{id}'}/events</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Accounts List Placeholder */}
      <div className="card p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Accounts</h3>
        <div className="text-center py-8">
          <div className="text-gray-400 mb-2">
            <Eye size={48} className="mx-auto" />
          </div>
          <p className="text-gray-500">No accounts to display</p>
          <p className="text-sm text-gray-400 mt-1">
            Create an account to see it listed here
          </p>
        </div>
      </div>
    </div>
  );
};

export default AccountManagement;