import React, { useState, useMemo } from 'react';
import type { EventResponse } from '../types';
import { Download, Eye, EyeOff, Search, Filter } from 'lucide-react';
import { exportTransactionsToCSV, TransactionRecord } from '../utils/csvExport';

interface CSVViewerProps {
  events: EventResponse[];
  accountId: string;
}

const CSVViewer: React.FC<CSVViewerProps> = ({ events, accountId }) => {
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [sortField, setSortField] = useState<keyof TransactionRecord>('date');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

  // Convert events to transaction records for display
  const transactionRecords: TransactionRecord[] = useMemo(() => {
    return events.map((event) => {
      const date = new Date(event.timestamp);
      const eventData = event.data || {};
      
      let type = '';
      let description = '';
      let amount = '';
      
      switch (event.event_type.toLowerCase()) {
        case 'bank_account_created':
        case 'bank_account_created_v1':
          type = 'Account Creation';
          description = `Account created for ${eventData.email || 'user'}`;
          amount = eventData.initial_balance 
            ? `${Number(eventData.initial_balance).toLocaleString('vi-VN')} VND` 
            : eventData.balance 
              ? `${Number(eventData.balance).toLocaleString('vi-VN')} VND`
              : '0 VND';
          break;
        case 'balance_deposited':
        case 'balance_deposited_v1':
          type = 'Deposit';
          description = `Money deposited to account${eventData.paymentID ? ` (Payment: ${eventData.paymentID})` : ''}`;
          amount = `+${Number(eventData.amount || 0).toLocaleString('vi-VN')} VND`;
          break;
        case 'balance_withdrawed':
        case 'balance_withdrawn':
        case 'balance_withdrawed_v1':
          type = 'Withdrawal';
          description = `Money withdrawn from account${eventData.paymentID ? ` (Payment: ${eventData.paymentID})` : ''}`;
          amount = `-${Number(eventData.amount || 0).toLocaleString('vi-VN')} VND`;
          break;
        default:
          type = 'Unknown';
          description = `Unknown transaction type: ${event.event_type}`;
          amount = '0 VND';
      }

      return {
        date: date.toLocaleDateString('vi-VN'),
        time: date.toLocaleTimeString('vi-VN'),
        type,
        description,
        amount,
        paymentId: eventData.paymentID || eventData.payment_id || 'N/A',
        version: event.version,
      };
    });
  }, [events]);

  // Filter and sort records
  const filteredAndSortedRecords = useMemo(() => {
    let filtered = transactionRecords;

    // Apply search filter
    if (searchTerm) {
      filtered = filtered.filter(record =>
        Object.values(record).some(value =>
          value.toString().toLowerCase().includes(searchTerm.toLowerCase())
        )
      );
    }

    // Apply type filter
    if (filterType !== 'all') {
      filtered = filtered.filter(record => 
        record.type.toLowerCase().includes(filterType.toLowerCase())
      );
    }

    // Apply sorting
    filtered.sort((a, b) => {
      const aValue = a[sortField];
      const bValue = b[sortField];
      
      let comparison = 0;
      if (aValue < bValue) comparison = -1;
      if (aValue > bValue) comparison = 1;
      
      return sortDirection === 'asc' ? comparison : -comparison;
    });

    return filtered;
  }, [transactionRecords, searchTerm, filterType, sortField, sortDirection]);

  const handleExportCSV = () => {
    exportTransactionsToCSV(events, accountId);
  };

  const handleSort = (field: keyof TransactionRecord) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const getSortIcon = (field: keyof TransactionRecord) => {
    if (sortField !== field) return null;
    return sortDirection === 'asc' ? '↑' : '↓';
  };

  const uniqueTypes = useMemo(() => {
    const types = Array.from(new Set(transactionRecords.map(r => r.type)));
    return types.sort();
  }, [transactionRecords]);

  return (
    <div className="space-y-4">
      {/* Control Buttons */}
      <div className="flex flex-wrap gap-2 justify-between items-center">
        <div className="flex gap-2">
          <button
            onClick={() => setIsViewerOpen(!isViewerOpen)}
            className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
          >
            {isViewerOpen ? <EyeOff className="w-4 h-4 mr-2" /> : <Eye className="w-4 h-4 mr-2" />}
            {isViewerOpen ? 'Hide CSV View' : 'View as CSV'}
          </button>
          
          <button
            onClick={handleExportCSV}
            className="inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors"
          >
            <Download className="w-4 h-4 mr-2" />
            Download CSV
          </button>
        </div>

        <div className="text-sm text-gray-600">
          {filteredAndSortedRecords.length} of {transactionRecords.length} records
        </div>
      </div>

      {/* CSV Viewer */}
      {isViewerOpen && (
        <div className="border border-gray-200 rounded-lg overflow-hidden bg-white">
          {/* Filters */}
          <div className="p-4 bg-gray-50 border-b flex flex-wrap gap-4 items-center">
            <div className="flex items-center gap-2">
              <Search className="w-4 h-4 text-gray-500" />
              <input
                type="text"
                placeholder="Search records..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="px-3 py-1 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            
            <div className="flex items-center gap-2">
              <Filter className="w-4 h-4 text-gray-500" />
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
                className="px-3 py-1 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">All Types</option>
                {uniqueTypes.map(type => (
                  <option key={type} value={type}>{type}</option>
                ))}
              </select>
            </div>
          </div>

          {/* Table */}
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-100">
                <tr>
                  <th 
                    className="px-4 py-2 text-left font-medium text-gray-700 cursor-pointer hover:bg-gray-200"
                    onClick={() => handleSort('date')}
                  >
                    Date {getSortIcon('date')}
                  </th>
                  <th 
                    className="px-4 py-2 text-left font-medium text-gray-700 cursor-pointer hover:bg-gray-200"
                    onClick={() => handleSort('time')}
                  >
                    Time {getSortIcon('time')}
                  </th>
                  <th 
                    className="px-4 py-2 text-left font-medium text-gray-700 cursor-pointer hover:bg-gray-200"
                    onClick={() => handleSort('type')}
                  >
                    Type {getSortIcon('type')}
                  </th>
                  <th className="px-4 py-2 text-left font-medium text-gray-700">
                    Description
                  </th>
                  <th 
                    className="px-4 py-2 text-right font-medium text-gray-700 cursor-pointer hover:bg-gray-200"
                    onClick={() => handleSort('amount')}
                  >
                    Amount {getSortIcon('amount')}
                  </th>
                  <th className="px-4 py-2 text-left font-medium text-gray-700">
                    Payment ID
                  </th>
                  <th 
                    className="px-4 py-2 text-center font-medium text-gray-700 cursor-pointer hover:bg-gray-200"
                    onClick={() => handleSort('version')}
                  >
                    Version {getSortIcon('version')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {filteredAndSortedRecords.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                      No records found
                    </td>
                  </tr>
                ) : (
                  filteredAndSortedRecords.map((record, index) => (
                    <tr key={index} className={`border-t hover:bg-gray-50 ${index % 2 === 0 ? 'bg-white' : 'bg-gray-50/30'}`}>
                      <td className="px-4 py-2 font-mono text-xs">{record.date}</td>
                      <td className="px-4 py-2 font-mono text-xs">{record.time}</td>
                      <td className="px-4 py-2">
                        <span className={`inline-flex px-2 py-1 text-xs rounded-full ${
                          record.type === 'Deposit' ? 'bg-green-100 text-green-800' :
                          record.type === 'Withdrawal' ? 'bg-red-100 text-red-800' :
                          record.type === 'Account Creation' ? 'bg-blue-100 text-blue-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {record.type}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-gray-700">{record.description}</td>
                      <td className={`px-4 py-2 text-right font-medium ${
                        record.amount.startsWith('+') ? 'text-green-600' :
                        record.amount.startsWith('-') ? 'text-red-600' :
                        'text-gray-600'
                      }`}>
                        {record.amount}
                      </td>
                      <td className="px-4 py-2 font-mono text-xs text-gray-600">{record.paymentId}</td>
                      <td className="px-4 py-2 text-center font-mono text-xs">{record.version}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Summary */}
          {filteredAndSortedRecords.length > 0 && (
            <div className="p-4 bg-gray-50 border-t">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div>
                  <span className="text-gray-600">Total Records:</span>
                  <span className="ml-2 font-medium">{filteredAndSortedRecords.length}</span>
                </div>
                <div>
                  <span className="text-gray-600">Deposits:</span>
                  <span className="ml-2 font-medium text-green-600">
                    {filteredAndSortedRecords.filter(r => r.type === 'Deposit').length}
                  </span>
                </div>
                <div>
                  <span className="text-gray-600">Withdrawals:</span>
                  <span className="ml-2 font-medium text-red-600">
                    {filteredAndSortedRecords.filter(r => r.type === 'Withdrawal').length}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default CSVViewer;