import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  getTransactions,
  getTransaction,
  type TransactionsParams,
} from "../services/transactions.service";

export const transactionKeys = {
  all: ["transactions"] as const,
  list: (params: TransactionsParams) => ["transactions", params] as const,
  detail: (id: string) => ["transactions", id] as const,
};

export const useTransactions = (params: TransactionsParams = {}) => {
  return useQuery({
    queryKey: transactionKeys.list(params),
    queryFn: () => getTransactions(params),
    placeholderData: keepPreviousData,
  });
};

export const useTransaction = (id: string) => {
  return useQuery({
    queryKey: transactionKeys.detail(id),
    queryFn: () => getTransaction(id),
    enabled: !!id,
  });
};
