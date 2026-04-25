import api from "./api";
import type { AuthResponse } from "../types";

export interface SignupPayload {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export const signup = async (payload: SignupPayload): Promise<AuthResponse> => {
  const { data } = await api.post<AuthResponse>("/auth/signup", payload);
  return data;
};

export const login = async (payload: LoginPayload): Promise<AuthResponse> => {
  const { data } = await api.post<AuthResponse>("/auth/login", payload);
  return data;
};
