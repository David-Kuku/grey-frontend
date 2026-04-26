import axios from "axios";
import { getToken, removeToken } from "../utils/auth";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL + "/api/v1",
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401 && !err.config?.url?.includes("/auth/")) {
      removeToken();
      window.location.href = "/login";
    }
    return Promise.reject(err);
  },
);

export default api;
