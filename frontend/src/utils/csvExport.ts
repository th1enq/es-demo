import type { EventResponse } from '../types';

export interface TransactionRecord {
  date: string;
  time: string;
  type: string;
  description: string;
  amount: string;
  paymentId: string;
  version: number;
}

export const exportTransactionsToCSV = (events: EventResponse[], accountId: string) => {
  // Convert events to transaction records
  const transactions: TransactionRecord[] = events.map((event) => {
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

  // Sort by date (newest first)
  transactions.sort((a, b) => {
    const dateA = new Date(`${a.date} ${a.time}`);
    const dateB = new Date(`${b.date} ${b.time}`);
    return dateB.getTime() - dateA.getTime();
  });

  // CSV headers
  const headers = [
    'Date',
    'Time',
    'Transaction Type',
    'Description',
    'Amount',
    'Payment ID',
    'Event Version',
    'Event ID'
  ];

  // Convert to CSV format
  const csvContent = [
    // Header row
    headers.join(','),
    // Data rows
    ...transactions.map(transaction => [
      `"${transaction.date}"`,
      `"${transaction.time}"`,
      `"${transaction.type}"`,
      `"${transaction.description}"`,
      `"${transaction.amount}"`,
      `"${transaction.paymentId}"`,
      transaction.version.toString(),
      `"${events.find(e => e.timestamp === new Date(`${transaction.date} ${transaction.time}`).toISOString())?.event_id || 'N/A'}"`
    ].join(','))
  ].join('\n');

  // Add BOM for proper UTF-8 encoding in Excel
  const BOM = '\uFEFF';
  const finalContent = BOM + csvContent;

  // Generate filename with timestamp
  const now = new Date();
  const timestamp = now.toISOString().slice(0, 19).replace(/:/g, '-');
  const filename = `transactions_${accountId.slice(0, 8)}_${timestamp}.csv`;

  // Create and download file
  const blob = new Blob([finalContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  
  if (link.download !== undefined) {
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }

  return {
    filename,
    recordCount: transactions.length,
    dateRange: transactions.length > 0 ? {
      from: transactions[transactions.length - 1].date,
      to: transactions[0].date
    } : null
  };
};