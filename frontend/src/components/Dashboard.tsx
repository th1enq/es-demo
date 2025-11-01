import React from 'react';
import { 
  TrendingUp, 
  DollarSign, 
  Users, 
  CreditCard,
  ArrowUpRight,
  ArrowDownRight,
  MoreHorizontal
} from 'lucide-react';
import { Line, Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend
);

interface StatCardProps {
  title: string;
  value: string;
  change: string;
  isPositive: boolean;
  icon: React.ReactNode;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, change, isPositive, icon }) => {
  return (
    <div className="card p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <p className="text-2xl font-semibold text-gray-900 mt-1">{value}</p>
          <div className="flex items-center mt-2">
            {isPositive ? (
              <ArrowUpRight className="text-green-500" size={16} />
            ) : (
              <ArrowDownRight className="text-red-500" size={16} />
            )}
            <span className={`text-sm font-medium ml-1 ${
              isPositive ? 'text-green-600' : 'text-red-600'
            }`}>
              {change}
            </span>
          </div>
        </div>
        <div className="p-3 bg-primary-50 rounded-lg">
          {icon}
        </div>
      </div>
    </div>
  );
};

const Dashboard: React.FC = () => {
  // Sample data - you can replace this with real API data
  const stats = [
    {
      title: 'Total Balance',
      value: '₫615,804,500',
      change: '+2.5%',
      isPositive: true,
      icon: <DollarSign className="text-primary-600" size={24} />,
    },
    {
      title: 'Monthly Income',
      value: '₫205,700,000',
      change: '+12.3%',
      isPositive: true,
      icon: <TrendingUp className="text-primary-600" size={24} />,
    },
    {
      title: 'Active Accounts',
      value: '12',
      change: '+3',
      isPositive: true,
      icon: <CreditCard className="text-primary-600" size={24} />,
    },
    {
      title: 'Total Customers',
      value: '1,248',
      change: '+5.2%',
      isPositive: true,
      icon: <Users className="text-primary-600" size={24} />,
    },
  ];

  // Sample chart data
  const cashFlowData = {
    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
    datasets: [
      {
        label: 'Income',
        data: [2000, 2500, 3000, 2800, 3200, 3500],
        borderColor: '#22c55e',
        backgroundColor: '#22c55e',
        tension: 0.4,
      },
      {
        label: 'Outcome',
        data: [1500, 1800, 2200, 2000, 2400, 2600],
        borderColor: '#3b82f6',
        backgroundColor: '#3b82f6',
        tension: 0.4,
      },
    ],
  };

  const chartOptions = {
    responsive: true,
    plugins: {
      legend: {
        position: 'bottom' as const,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
      },
    },
  };

  const invoiceData = {
    labels: ['Jan 01-08', 'Jan 09-16', 'Jan 17-24', 'Jan 25-31', 'Feb 01-07'],
    datasets: [
      {
        data: [600, 800, 1200, 900, 1100],
        backgroundColor: '#3b82f6',
        borderRadius: 4,
      },
    ],
  };

  const barOptions = {
    responsive: true,
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
      },
    },
  };

  return (
    <div className="p-6 space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Cash Flow Chart */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900">Total cash flow</h3>
            <button className="p-2 hover:bg-gray-100 rounded-lg">
              <MoreHorizontal size={16} className="text-gray-400" />
            </button>
          </div>
          <div className="h-64">
            <Line data={cashFlowData} options={chartOptions} />
          </div>
        </div>

        {/* Invoice Chart */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900">Invoices owed to you</h3>
            <button className="text-sm text-primary-600 hover:text-primary-700 font-medium">
              New Sales Invoice
            </button>
          </div>
          <div className="h-64">
            <Bar data={invoiceData} options={barOptions} />
          </div>
        </div>
      </div>

      {/* Account Watchlist */}
      <div className="card p-6">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-semibold text-gray-900">Account watchlist</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left py-3 px-4 font-medium text-gray-500">Account</th>
                <th className="text-right py-3 px-4 font-medium text-gray-500">This Month</th>
                <th className="text-right py-3 px-4 font-medium text-gray-500">YTD</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b border-gray-100">
                <td className="py-3 px-4 text-gray-900">Sales</td>
                <td className="py-3 px-4 text-right text-gray-900">₫29,864,500</td>
                <td className="py-3 px-4 text-right text-gray-900">₫286,457,250</td>
              </tr>
              <tr className="border-b border-gray-100">
                <td className="py-3 px-4 text-gray-900">Advertising</td>
                <td className="py-3 px-4 text-right text-gray-900">₫172,975,500</td>
                <td className="py-3 px-4 text-right text-gray-900">₫232,784,000</td>
              </tr>
              <tr className="border-b border-gray-100">
                <td className="py-3 px-4 text-gray-900">Inventory</td>
                <td className="py-3 px-4 text-right text-gray-900">₫117,806,500</td>
                <td className="py-3 px-4 text-right text-gray-900">₫245,202,250</td>
              </tr>
              <tr className="border-b border-gray-100">
                <td className="py-3 px-4 text-gray-900">Entertainment</td>
                <td className="py-3 px-4 text-right text-gray-900">₫0</td>
                <td className="py-3 px-4 text-right text-gray-900">₫0</td>
              </tr>
              <tr>
                <td className="py-3 px-4 text-gray-900">Product</td>
                <td className="py-3 px-4 text-right text-gray-900">₫116,802,500</td>
                <td className="py-3 px-4 text-right text-gray-900">₫63,497,500</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;