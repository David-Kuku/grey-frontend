import { useState } from "react";
import { useTransactions } from "../../queries/transaction.queries";

export const PAGE_SIZE_OPTIONS = [5, 10, 20, 50] as const;

export const useTransactionsView = () => {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState<number>(20);
  const { data, isLoading, isFetching } = useTransactions({
    page,
    page_size: limit,
  });

  const handleLimitChange = (newLimit: number) => {
    setLimit(newLimit);
    setPage(1);
  };

  return {
    data,
    isLoading,
    isFetching,
    page,
    setPage,
    limit,
    handleLimitChange,
  };
};
