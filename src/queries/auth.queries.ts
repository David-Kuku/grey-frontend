import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  login,
  signup,
  type LoginPayload,
  type SignupPayload,
} from "../services/auth.service";
import { setToken, removeToken } from "../utils/auth";

export const useLogin = () => {
  return useMutation({
    mutationFn: (payload: LoginPayload) => login(payload),
    onSuccess: (data) => {
      setToken(data.token);
    },
  });
};

export const useSignup = () => {
  return useMutation({
    mutationFn: (payload: SignupPayload) => signup(payload),
    onSuccess: (data) => {
      setToken(data.token);
    },
  });
};

export const useLogout = () => {
  const qc = useQueryClient();
  return () => {
    removeToken();
    qc.clear();
    window.location.href = "/login";
  };
};
