export interface BankAccount {
  aggregateID: string;
  email: string;
  firstName: string;  // Response uses firstName
  lastName: string;   // Response uses lastName
  balance: {
    amount: number;
    currency: string;
  };
  updated_at?: string;
}

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

export interface CreateBankAccountRequest {
  email: string;
  first_name: string;
  last_name: string;
  balance: number;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}

export interface RegisterRequest {
  id?: string;
  email: string;
  password: string;
  first_name: string;    // Request uses first_name
  last_name: string;     // Request uses last_name
  initial_balance: number; // Request uses initial_balance
}

export interface RegisterResponse {
  user_id: string;
  email: string;
  message: string;
}

export interface DepositRequest {
  amount: number;
  payment_id: string;
}

export interface WithdrawRequest {
  amount: number;
  payment_id: string;
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