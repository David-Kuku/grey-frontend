import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createPayout, type PayoutPayload } from "../services/payout.service";
import { walletKeys } from "./wallet.queries";
import { transactionKeys } from "./transaction.queries";

export const payoutKeys = {
  detail: (id: string) => ["payout", id] as const,
};

export const useCreatePayout = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: PayoutPayload) => {
      return createPayout(payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: walletKeys.all });
      qc.invalidateQueries({ queryKey: transactionKeys.all });
    },
  });
};
