import React, { useState, useEffect } from 'react';
import { BankAccountService } from '../services/api';
import type { EventsHistoryResponse, EventResponse } from '../types';
import { Activity, Database, RefreshCw, Eye, Clock, Download } from 'lucide-react';
import { exportTransactionsToCSV } from '../utils/csvExport';

const Events: React.FC = () => {
  const [events, setEvents] = useState<EventsHistoryResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<EventResponse | null>(null);
  const [isExporting, setIsExporting] = useState(false);

  useEffect(() => {
    loadEvents();
    
    // Auto-reload every 5 seconds for real-time updates
    const interval = setInterval(() => {
      loadEvents();
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  const handleExportCSV = () => {
    if (!events || events.events.length === 0) {
      return;
    }

    try {
      setIsExporting(true);
      exportTransactionsToCSV(events.events, events.aggregate_id);
      
      // Close any open event details
      setTimeout(() => {
        setSelectedEvent(null);
      }, 100);
    } catch (error) {
      console.error('Export error:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const loadEvents = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const savedUser = localStorage.getItem('user');
      if (!savedUser) {
        setError('User information not found');
        return;
      }

      const user = JSON.parse(savedUser);
      const response = await BankAccountService.getEventsHistory(user.id);
      
      if (response.success && response.data) {
        setEvents(response.data);
      } else {
        setError(response.error?.message || 'Failed to load events history');
      }
    } catch (err: any) {
      console.error('Error loading events:', err);
      setError(err.response?.data?.error?.message || 'Failed to load events history');
    } finally {
      setLoading(false);
    }
  };

  const getEventTypeColor = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'bank_account_created':
      case 'bank_account_created_v1':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'balance_deposited':
      case 'balance_deposited_v1':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'balance_withdrawed':
      case 'balance_withdrawed_v1':
      case 'balance_withdrawn':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getEventIcon = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'bank_account_created':
      case 'bank_account_created_v1':
        return <Database className="h-4 w-4 text-blue-600" />;
      case 'balance_deposited':
      case 'balance_deposited_v1':
        return <div className="h-4 w-4 bg-green-600 rounded-full flex items-center justify-center">
          <span className="text-white font-bold text-xs">+</span>
        </div>;
      case 'balance_withdrawed':
      case 'balance_withdrawed_v1':
      case 'balance_withdrawn':
        return <div className="h-4 w-4 bg-red-600 rounded-full flex items-center justify-center">
          <span className="text-white font-bold text-xs">-</span>
        </div>;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const getEventDescription = (eventType: string, data: any) => {
    switch (eventType.toLowerCase()) {
      case 'bank_account_created':
      case 'bank_account_created_v1':
        return `Account created for ${data.email || 'user'}`;
      case 'balance_deposited':
      case 'balance_deposited_v1':
        return `Deposited ${data.amount?.toLocaleString('vi-VN') || 'N/A'} VND`;
      case 'balance_withdrawed':
      case 'balance_withdrawn':
      case 'balance_withdrawed_v1':
        return `Withdrew ${data.amount?.toLocaleString('vi-VN') || 'N/A'} VND`;
      default:
        return 'Unknown event';
    }
  };

  const getEventRowColor = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'bank_account_created':
      case 'bank_account_created_v1':
        return 'hover:bg-blue-50';
      case 'balance_deposited':
      case 'balance_deposited_v1':
        return 'hover:bg-green-50';
      case 'balance_withdrawed':
      case 'balance_withdrawed_v1':
      case 'balance_withdrawn':
        return 'hover:bg-red-50';
      default:
        return 'hover:bg-gray-50';
    }
  };

  const formatEventType = (eventType: string) => {
    return eventType
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
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
            onClick={loadEvents}
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
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 flex items-center space-x-2">
            <Activity className="h-7 w-7 text-indigo-600" />
            <span>Events History</span>
          </h2>
          <p className="text-gray-600">View all events and transactions for your account</p>
        </div>
        
        <div className="flex items-center space-x-3">
          {/* Export CSV Button */}
          {events && events.events.length > 0 && (
            <button
              onClick={handleExportCSV}
              disabled={isExporting}
              className="flex items-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50"
            >
              <Download className={`h-4 w-4 ${isExporting ? 'animate-pulse' : ''}`} />
              <span>{isExporting ? 'Exporting...' : 'Export CSV'}</span>
            </button>
          )}
          
          <button
            onClick={loadEvents}
            disabled={loading}
            className="flex items-center space-x-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {events ? (
        <>
          {/* Summary */}
          <div className="mb-6 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-blue-100 rounded-lg">
                  <Database className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Total Events</p>
                  <p className="text-2xl font-bold text-gray-900">{events.total_events}</p>
                </div>
              </div>
              
              {/* Event Type Statistics */}
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-blue-100 rounded-lg">
                  <Database className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Account Created</p>
                  <p className="text-2xl font-bold text-blue-600">
                    {events.events.filter(e => 
                      e.event_type.toLowerCase().includes('account_created')
                    ).length}
                  </p>
                </div>
              </div>
              
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-green-100 rounded-lg">
                  <div className="h-6 w-6 bg-green-600 rounded-full flex items-center justify-center">
                    <span className="text-white font-bold text-sm">+</span>
                  </div>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Deposits</p>
                  <p className="text-2xl font-bold text-green-600">
                    {events.events.filter(e => 
                      e.event_type.toLowerCase().includes('deposited')
                    ).length}
                  </p>
                </div>
              </div>
              
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-red-100 rounded-lg">
                  <div className="h-6 w-6 bg-red-600 rounded-full flex items-center justify-center">
                    <span className="text-white font-bold text-sm">-</span>
                  </div>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Withdrawals</p>
                  <p className="text-2xl font-bold text-red-600">
                    {events.events.filter(e => 
                      e.event_type.toLowerCase().includes('withdraw')
                    ).length}
                  </p>
                </div>
              </div>
            </div>
            
            {/* Additional Info */}
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <p className="text-sm font-medium text-gray-500">Account ID</p>
                  <p className="text-sm font-mono text-gray-900 truncate">{events.aggregate_id}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Last Updated</p>
                  <p className="text-sm text-gray-900">
                    {events.events.length > 0 
                      ? formatTimestamp(events.events[events.events.length - 1].timestamp)
                      : 'No events'
                    }
                  </p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Export Options</p>
                  <button
                    onClick={handleExportCSV}
                    disabled={isExporting || events.events.length === 0}
                    className="text-sm px-3 py-1 bg-green-100 text-green-700 rounded-full hover:bg-green-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isExporting ? 'Exporting...' : 'Download CSV'}
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Events List */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">Event Timeline</h3>
                
                {/* Legend */}
                <div className="flex items-center space-x-4 text-xs">
                  <div className="flex items-center space-x-1">
                    <div className="w-3 h-3 bg-blue-100 border border-blue-200 rounded"></div>
                    <span className="text-gray-600">Account Created</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-3 h-3 bg-green-100 border border-green-200 rounded"></div>
                    <span className="text-gray-600">Deposit</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-3 h-3 bg-red-100 border border-red-200 rounded"></div>
                    <span className="text-gray-600">Withdrawal</span>
                  </div>
                </div>
              </div>
            </div>

            {events.events.length > 0 ? (
              <div className="divide-y divide-gray-200">
                {events.events.map((event) => (
                  <div key={event.event_id} className={`p-6 transition-colors ${getEventRowColor(event.event_type)}`}>
                    <div className="flex items-start space-x-4">
                      {/* Event Icon */}
                      <div className={`flex-shrink-0 p-2 rounded-lg border ${getEventTypeColor(event.event_type)}`}>
                        {getEventIcon(event.event_type)}
                      </div>

                      {/* Event Details */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="text-lg font-medium text-gray-900">
                              {formatEventType(event.event_type)}
                            </h4>
                            <p className="text-sm text-gray-600 mt-1">
                              {getEventDescription(event.event_type, event.data)}
                            </p>
                            <div className="flex items-center space-x-4 mt-1">
                              <span className="text-sm text-gray-500 flex items-center space-x-1">
                                <Clock className="h-3 w-3" />
                                <span>{formatTimestamp(event.timestamp)}</span>
                              </span>
                              <span className="text-sm text-gray-500">
                                Version: {event.version}
                              </span>
                            </div>
                          </div>
                          <button
                            onClick={() => setSelectedEvent(selectedEvent?.event_id === event.event_id ? null : event)}
                            className="flex items-center space-x-1 px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                          >
                            <Eye className="h-3 w-3" />
                            <span>{selectedEvent?.event_id === event.event_id ? 'Hide' : 'View'} Data</span>
                          </button>
                        </div>

                        {/* Event Data */}
                        {selectedEvent?.event_id === event.event_id && (
                          <div className="mt-4 p-4 bg-gray-50 rounded-lg">
                            <h5 className="text-sm font-medium text-gray-700 mb-2">Event Data:</h5>
                            <pre className="text-xs bg-white p-3 rounded border overflow-x-auto">
                              {JSON.stringify(event.data, null, 2)}
                            </pre>
                            {event.metadata && (
                              <>
                                <h5 className="text-sm font-medium text-gray-700 mb-2 mt-3">Metadata:</h5>
                                <pre className="text-xs bg-white p-3 rounded border overflow-x-auto">
                                  {JSON.stringify(event.metadata, null, 2)}
                                </pre>
                              </>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="p-12 text-center">
                <Activity className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 mb-2">No Events Found</h3>
                <p className="text-gray-500">
                  No events have been recorded for this account yet.
                </p>
              </div>
            )}
          </div>
        </>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <Activity className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Events Data</h3>
          <p className="text-gray-500">Unable to load events history.</p>
        </div>
      )}
    </div>
  );
};

export default Events;