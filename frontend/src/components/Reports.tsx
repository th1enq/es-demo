import React, { useState, useEffect } from 'react';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { BankAccountService } from '../services/api';
import { TrendingUp, AlertCircle, RefreshCw } from 'lucide-react';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

interface VersionData {
  version: number;
  balance: number;
  timestamp?: string;
}

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

const Reports: React.FC = () => {
  const [versionData, setVersionData] = useState<VersionData[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [maxVersion, setMaxVersion] = useState(20);

  // Get current user from localStorage
  useEffect(() => {
    const savedUser = localStorage.getItem('user');
    if (savedUser) {
      try {
        const parsedUser = JSON.parse(savedUser);
        setUser(parsedUser);
        // Auto-load data when component mounts
        fetchVersionData(parsedUser.id);
      } catch (error) {
        setError('Không thể lấy thông tin người dùng');
      }
    } else {
      setError('Vui lòng đăng nhập để xem báo cáo');
    }
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND'
    }).format(amount);
  };

  const fetchVersionData = async (accountId?: string) => {
    const targetAccountId = accountId || user?.id;
    
    if (!targetAccountId) {
      setError('Không có thông tin tài khoản');
      return;
    }

    setLoading(true);
    setError('');
    setVersionData([]);

    try {
      // First, get current account to check if it exists
      const currentAccount = await BankAccountService.getAccount(targetAccountId);
      if (!currentAccount.success || !currentAccount.data) {
        throw new Error('Không tìm thấy tài khoản');
      }

      // Get events to determine actual max version
      const eventsResponse = await BankAccountService.getEventsHistory(targetAccountId);
      const actualMaxVersion = eventsResponse.data?.events?.length || maxVersion;
      
      const promises: Promise<VersionData | null>[] = [];
      
      // Fetch data for each version from 1 to actual max version
      for (let version = 1; version <= Math.min(actualMaxVersion, maxVersion); version++) {
        promises.push(
          BankAccountService.getAccountByVersion(targetAccountId, version)
            .then(response => {
              if (response.success && response.data) {
                return {
                  version,
                  balance: response.data.balance?.amount || 0,
                  timestamp: response.data.updated_at
                };
              }
              return null;
            })
            .catch(() => null) // Skip failed versions
        );
      }

      const results = await Promise.all(promises);
      const validData = results.filter((data): data is VersionData => data !== null);
      
      if (validData.length === 0) {
        throw new Error('Không có dữ liệu version nào được tìm thấy');
      }

      setVersionData(validData);
    } catch (err: any) {
      setError(err.message || 'Có lỗi xảy ra khi tải dữ liệu');
      setVersionData([]);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    if (user?.id) {
      fetchVersionData(user.id);
    }
  };

  const chartData = {
    labels: versionData.map(d => `V${d.version}`),
    datasets: [
      {
        label: 'Số dư tài khoản',
        data: versionData.map(d => d.balance),
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        borderWidth: 2,
        fill: true,
        tension: 0.4,
        pointBackgroundColor: '#3b82f6',
        pointBorderColor: '#ffffff',
        pointBorderWidth: 2,
        pointRadius: 6,
        pointHoverRadius: 8,
      },
    ],
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top' as const,
        labels: {
          font: {
            size: 14,
          },
        },
      },
      title: {
        display: true,
        text: `Biểu đồ số dư theo Version - ${user?.first_name} ${user?.last_name}`,
        font: {
          size: 16,
          weight: 'bold' as const,
        },
      },
      tooltip: {
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        titleColor: '#ffffff',
        bodyColor: '#ffffff',
        borderColor: '#3b82f6',
        borderWidth: 1,
        callbacks: {
          label: function(context: any) {
            return `Số dư: ${formatCurrency(context.parsed.y)}`;
          },
        },
      },
    },
    scales: {
      x: {
        title: {
          display: true,
          text: 'Version',
          font: {
            size: 14,
            weight: 'bold' as const,
          },
        },
        grid: {
          color: 'rgba(0, 0, 0, 0.1)',
        },
      },
      y: {
        title: {
          display: true,
          text: 'Số dư (VND)',
          font: {
            size: 14,
            weight: 'bold' as const,
          },
        },
        grid: {
          color: 'rgba(0, 0, 0, 0.1)',
        },
        ticks: {
          callback: function(value: any) {
            return formatCurrency(value);
          },
        },
      },
    },
    interaction: {
      intersect: false,
      mode: 'index' as const,
    },
  };

  // Calculate statistics
  const stats = versionData.length > 0 ? {
    totalVersions: versionData.length,
    currentBalance: versionData[versionData.length - 1]?.balance || 0,
    initialBalance: versionData[0]?.balance || 0,
    maxBalance: Math.max(...versionData.map(d => d.balance)),
    minBalance: Math.min(...versionData.map(d => d.balance)),
    change: versionData.length > 1 ? versionData[versionData.length - 1].balance - versionData[0].balance : 0
  } : null;

  return (
    <div className="p-6 space-y-6">
      <div className="card p-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-semibold text-gray-900 flex items-center">
            <TrendingUp className="mr-3 text-primary-600" size={28} />
            Báo cáo Số dư theo Version
          </h2>
          
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <label htmlFor="maxVersion" className="text-sm font-medium text-gray-700">
                Max Version:
              </label>
              <input
                type="number"
                id="maxVersion"
                value={maxVersion}
                onChange={(e) => setMaxVersion(parseInt(e.target.value) || 20)}
                min="1"
                max="100"
                className="w-20 px-2 py-1 border border-gray-300 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                disabled={loading}
              />
            </div>
            
            <button
              onClick={handleRefresh}
              disabled={loading || !user}
              className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center"
            >
              {loading ? (
                <RefreshCw className="animate-spin" size={20} />
              ) : (
                <>
                  <RefreshCw className="mr-2" size={20} />
                  Làm mới
                </>
              )}
            </button>
          </div>
        </div>

        {/* User Info */}
        {user && (
          <div className="bg-blue-50 p-4 rounded-lg mb-6">
            <p className="text-sm text-blue-600 font-medium">Tài khoản hiện tại</p>
            <p className="text-lg font-bold text-blue-900">{user.first_name} {user.last_name}</p>
            <p className="text-sm text-blue-700">{user.email}</p>
            <p className="text-xs text-blue-600 mt-1">ID: {user.id}</p>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center">
            <AlertCircle className="text-red-500 mr-3" size={20} />
            <span className="text-red-700">{error}</span>
          </div>
        )}

        {/* Statistics */}
        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-6">
            <div className="bg-blue-50 p-4 rounded-lg">
              <p className="text-sm text-blue-600 font-medium">Tổng Versions</p>
              <p className="text-2xl font-bold text-blue-900">{stats.totalVersions}</p>
            </div>
            <div className="bg-green-50 p-4 rounded-lg">
              <p className="text-sm text-green-600 font-medium">Số dư hiện tại</p>
              <p className="text-lg font-bold text-green-900">{formatCurrency(stats.currentBalance)}</p>
            </div>
            <div className="bg-purple-50 p-4 rounded-lg">
              <p className="text-sm text-purple-600 font-medium">Số dư ban đầu</p>
              <p className="text-lg font-bold text-purple-900">{formatCurrency(stats.initialBalance)}</p>
            </div>
            <div className="bg-orange-50 p-4 rounded-lg">
              <p className="text-sm text-orange-600 font-medium">Số dù cao nhất</p>
              <p className="text-lg font-bold text-orange-900">{formatCurrency(stats.maxBalance)}</p>
            </div>
            <div className="bg-red-50 p-4 rounded-lg">
              <p className="text-sm text-red-600 font-medium">Số dư thấp nhất</p>
              <p className="text-lg font-bold text-red-900">{formatCurrency(stats.minBalance)}</p>
            </div>
            <div className={`p-4 rounded-lg ${stats.change >= 0 ? 'bg-green-50' : 'bg-red-50'}`}>
              <p className={`text-sm font-medium ${stats.change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                Thay đổi
              </p>
              <p className={`text-lg font-bold ${stats.change >= 0 ? 'text-green-900' : 'text-red-900'}`}>
                {stats.change >= 0 ? '+' : ''}{formatCurrency(stats.change)}
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Chart */}
      {loading && (
        <div className="card p-6">
          <div className="text-center py-12">
            <RefreshCw className="animate-spin mx-auto text-primary-500 mb-4" size={48} />
            <p className="text-gray-500 text-lg">Đang tải dữ liệu...</p>
          </div>
        </div>
      )}

      {!loading && versionData.length > 0 && (
        <div className="card p-6">
          <div className="h-96">
            <Line data={chartData} options={chartOptions} />
          </div>
        </div>
      )}

      {!loading && versionData.length === 0 && !error && user && (
        <div className="card p-6">
          <div className="text-center py-12">
            <AlertCircle className="mx-auto text-gray-400 mb-4" size={48} />
            <p className="text-gray-500 text-lg">Chưa có dữ liệu để hiển thị</p>
            <p className="text-gray-400">Vui lòng thực hiện một số giao dịch để xem biểu đồ</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default Reports;