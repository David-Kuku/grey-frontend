import api from "./api";
import type { Wallet } from "../types";

export const getWallet = async (): Promise<Wallet> => {
  const { data } = await api.get<Wallet>("/wallet");
  return data;
};
