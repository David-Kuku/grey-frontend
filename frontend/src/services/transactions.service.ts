import api from "./api";
import type { PaginatedTransactions, Transaction } from "../types";

export interface TransactionsParams {
  page?: number;
  limit?: number;
}

export const getTransactions = async (
  params: TransactionsParams = {},
): Promise<PaginatedTransactions> => {
  const { data } = await api.get<PaginatedTransactions>("/transactions", {
    params,
  });
  return data;
};

export const getTransaction = async (id: string): Promise<Transaction> => {
  const { data } = await api.get<Transaction>(`/transactions/${id}`);
  return data;
};
