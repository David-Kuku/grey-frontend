import api from "./api";
import type { PaginatedTransactions, TransactionSummary } from "../types";

export interface TransactionsParams {
  page?: number;
  page_size?: number;
}

export const getTransactions = async (
  params: TransactionsParams = {},
): Promise<PaginatedTransactions> => {
  const { data } = await api.get<PaginatedTransactions>("/transactions", {
    params,
  });
  return data;
};

export const getTransaction = async (id: string): Promise<TransactionSummary> => {
  const { data } = await api.get<TransactionSummary>(`/transactions/${id}`);
  return data;
};
