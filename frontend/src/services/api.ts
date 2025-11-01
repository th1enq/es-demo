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

// Request interceptor
api.interceptors.request.use(
  (config) => {
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

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