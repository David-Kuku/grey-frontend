import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createPayout,
  getPayout,
  type PayoutPayload,
} from "../services/payout.service";
import { walletKeys } from "./wallet.queries";
import { transactionKeys } from "./transaction.queries";

export const payoutKeys = {
  detail: (id: string) => ["payout", id] as const,
};

export const useCreatePayout = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: PayoutPayload) => createPayout(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: walletKeys.all });
      qc.invalidateQueries({ queryKey: transactionKeys.all });
    },
  });
};

export const usePayout = (id: string) => {
  return useQuery({
    queryKey: payoutKeys.detail(id),
    queryFn: () => getPayout(id),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === "pending" || status === "processing" ? 3000 : false;
    },
  });
};
