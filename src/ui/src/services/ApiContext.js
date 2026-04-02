import React, { createContext, useState, useEffect } from 'react';
import axios from 'axios';

const API_URL =
  window.localStorage.getItem('apiUrl') ||
  process.env.REACT_APP_API_URL ||
  'http://localhost:8000';

const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' }
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
