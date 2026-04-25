import { useState } from "react";
import { useTransactions } from "../../queries/transaction.queries";

const PAGE_SIZE = 20;

export const useTransactionsView = () => {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useTransactions({ page, limit: PAGE_SIZE });

  return { data, isLoading, page, setPage };
};
