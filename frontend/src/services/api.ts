import axios from 'axios';
import type { 
  APIResponse, 
  BankAccount, 
  CreateBankAccountRequest, 
  DepositRequest, 
  WithdrawRequest,
  EventsHistoryResponse 
} from '../types';

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid, redirect to login
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

export class AuthService {
  static async login(email: string, password: string): Promise<APIResponse> {
    const response = await api.post('/auth/login', { email, password });
    return response.data;
  }

  static async register(data: {
    email: string;
    password: string;
    firstName: string;
    lastName: string;
    initialBalance: number;
  }): Promise<APIResponse> {
    // Generate UUID for the account - fallback for environments without crypto.randomUUID
    const generateUUID = () => {
      if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        return crypto.randomUUID();
      }
      // Fallback UUID generation
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
      });
    };
    
    // Map frontend field names to backend field names
    const requestData = {
      id: generateUUID(), // Generate UUID for the account
      email: data.email,
      password: data.password,
      first_name: data.firstName,    // Map to first_name
      last_name: data.lastName,      // Map to last_name
      initial_balance: data.initialBalance, // Map to initial_balance
    };
    const response = await api.post('/auth/register', requestData);
    return response.data;
  }

  static async refreshToken(refreshToken: string): Promise<APIResponse> {
    const response = await api.post('/auth/refresh', { refresh_token: refreshToken });
    return response.data;
  }

  static async logout(): Promise<APIResponse> {
    const response = await api.post('/auth/logout');
    return response.data;
  }
}

export class BankAccountService {
  static async createAccount(data: CreateBankAccountRequest): Promise<APIResponse> {
    const response = await api.post('/bank_accounts', data);
    return response.data;
  }

  static async getAccount(id: string, fromEventStore = false): Promise<APIResponse<BankAccount>> {
    const response = await api.get(`/bank_accounts/${id}`, {
      params: { from_event_store: fromEventStore }
    });
    return response.data;
  }

  static async deposit(id: string, data: DepositRequest): Promise<APIResponse> {
    const response = await api.post(`/bank_accounts/${id}/deposite`, data);
    return response.data;
  }

  static async withdraw(id: string, data: WithdrawRequest): Promise<APIResponse> {
    const response = await api.post(`/bank_accounts/${id}/withdraw`, data);
    return response.data;
  }

  static async getEventsHistory(id: string): Promise<APIResponse<EventsHistoryResponse>> {
    const response = await api.get(`/bank_accounts/${id}/events`);
    return response.data;
  }
}

export default api;
export { api };