import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
    baseURL: API_URL,
});

// Automatically attach JWT token to every request
api.interceptors.request.use((config) => {
    if (typeof window !== 'undefined') {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
    }
    return config;
});

// Redirect to login on 401
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401 && typeof window !== 'undefined') {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

// Auth
export const authApi = {
    register: (data: { name: string; email: string; password: string }) =>
        api.post('/auth/register', data),
    login: (data: { email: string; password: string }) =>
        api.post('/auth/login', data),
};

// Accounts
export const accountsApi = {
    getAll: () => api.get('/api/accounts'),
    create: (data: object) => api.post('/api/accounts', data),
    delete: (id: string) => api.delete(`/api/accounts/${id}`),
};

// Transactions
export const transactionsApi = {
    getByAccount: (accountId: string, params?: object) =>
        api.get(`/api/accounts/${accountId}/transactions`, { params }),
    create: (accountId: string, data: object) =>
        api.post(`/api/accounts/${accountId}/transactions`, data),
    getSummary: (params?: object) =>
        api.get('/api/transactions/summary', { params }),
};

// Budgets
export const budgetsApi = {
    getAll: () => api.get('/api/budgets'),
    create: (data: object) => api.post('/api/budgets', data),
    delete: (id: string) => api.delete(`/api/budgets/${id}`),
    getAlerts: () => api.get('/api/budgets/alerts'),
};

// Insights
export const insightsApi = {
    getAll: () => api.get('/api/insights'),
    generateWeekly: () => api.post('/api/insights/weekly'),
    generateBudget: () => api.post('/api/insights/budget'),
};