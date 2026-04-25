import { useWallet } from "../../queries/wallet.queries";
import { useTransactions } from "../../queries/transaction.queries";

export const useDashboardView = () => {
  const { data: wallet, isLoading: walletLoading } = useWallet();
  const { data: txns, isLoading: txnsLoading } = useTransactions({
    page_size: 5,
  });

  return { wallet, walletLoading, txns, txnsLoading };
};
