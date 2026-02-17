import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

// Create axios instance with default config
const api = axios.create({
  baseURL: API_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Identification API
export const identifyPlant = async (file) => {
  const formData = new FormData();
  formData.append('image', file);

  const response = await api.post('/identify', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });

  return response.data;
};

// Chat API
export const sendChatMessage = async (identificationId, message) => {
  const response = await api.post('/chat', {
    identification_id: identificationId,
    message: message,
  });

  return response.data;
};

export const getChatHistory = async (identificationId) => {
  const response = await api.get(`/chat/${identificationId}`);
  return response.data;
};

// History API
export const getHistory = async (limit = 20, offset = 0) => {
  const response = await api.get('/history', {
    params: { limit, offset },
  });
  return response.data;
};

export const getHistoryById = async (id) => {
  const response = await api.get(`/history/${id}`);
  return response.data;
};

export const getHistoryWithChat = async (id) => {
  const response = await api.get(`/history/${id}/with-chat`);
  return response.data;
};

export default api;
