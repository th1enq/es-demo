export interface BankAccount {
  aggregateID: string;
  email: string;
  firstName: string;
  lastName: string;
  balance: {
    amount: number;
    currency: string;
  };
  status: string;
}

export interface CreateBankAccountRequest {
  email: string;
  first_name: string;
  last_name: string;
  balance: number;
  status?: string;
  password: string;
}

export interface DepositRequest {
  amount: number;
}

export interface WithdrawRequest {
  amount: number;
}

export interface APIResponse<T = any> {
  success: boolean;
  code: string;
  message: string;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
}

export interface EventResponse {
  event_id: string;
  aggregate_id: string;
  event_type: string;
  aggregate_type: string;
  version: number;
  data: any;
  metadata?: any;
  timestamp: string;
}

export interface EventsHistoryResponse {
  aggregate_id: string;
  total_events: number;
  events: EventResponse[];
}