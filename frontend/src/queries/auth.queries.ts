import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  login,
  signup,
  type LoginPayload,
  type SignupPayload,
} from "../services/auth.service";
import { setToken, removeToken } from "../utils/auth";
import { useAuthStore } from "../store/authStore";

export const useLogin = () => {
  const setUser = useAuthStore((s) => s.setUser);
  return useMutation({
    mutationFn: (payload: LoginPayload) => login(payload),
    onSuccess: (data) => {
      setToken(data.token);
      setUser(data.user);
    },
  });
};

export const useSignup = () => {
  const setUser = useAuthStore((s) => s.setUser);
  return useMutation({
    mutationFn: (payload: SignupPayload) => signup(payload),
    onSuccess: (data) => {
      setToken(data.token);
      setUser(data.user);
    },
  });
};

export const useLogout = () => {
  const qc = useQueryClient();
  const clearUser = useAuthStore((s) => s.clearUser);
  return () => {
    removeToken();
    clearUser();
    qc.clear();
    window.location.href = "/login";
  };
};
