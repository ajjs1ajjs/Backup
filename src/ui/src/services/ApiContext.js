import React, { createContext, useState, useEffect } from 'react';
import axios from 'axios';

const API_URL =
  window.localStorage.getItem('apiUrl') ||
  process.env.REACT_APP_API_URL ||
  '';

const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' }
});

api.interceptors.request.use((config) => {
  try {
    const raw = localStorage.getItem('auth-storage');
    if (raw) {
      const parsed = JSON.parse(raw);
      const token = parsed?.state?.token;
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
  } catch (e) { /* ignore parse errors */ }
  return config;
});

api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const fetchWithAuth = async (url, options = {}) => {
  const requestUrl = /^https?:\/\//i.test(url)
    ? url
    : `${API_URL}${url}`;
  const headers = { 'Content-Type': 'application/json', ...options.headers };
  try {
    const raw = localStorage.getItem('auth-storage');
    if (raw) {
      const parsed = JSON.parse(raw);
      const token = parsed?.state?.token;
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
    }
  } catch (e) { /* ignore */ }
  const response = await fetch(requestUrl, { ...options, headers });
  if (response.status === 401) {
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }
  return response;
};

export const useApi = (url) => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const refetch = async () => {
    setLoading(true);
    try {
      const response = await api.get(url);
      setData(response.data);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [url]);

  return { data, loading, error, refetch };
};

export const useApiMutation = (url, method = 'POST') => {
  const [loading, setLoading] = useState(false);

  const mutate = async (body) => {
    setLoading(true);
    try {
      const response = await api({ url, method, data: body });
      return response.data;
    } finally {
      setLoading(false);
    }
  };

  return { mutate, loading };
};

export const ApiContext = createContext({
  api,
  useApi,
  useApiMutation
});

export const ApiProvider = ({ children }) => {
  return (
    <ApiContext.Provider value={{ api, useApi, useApiMutation }}>
      {children}
    </ApiContext.Provider>
  );
};
