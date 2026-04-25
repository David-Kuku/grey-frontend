import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getQuote,
  executeConversion,
  type QuotePayload,
} from "../services/conversion.service";
import { walletKeys } from "./wallet.queries";
import { transactionKeys } from "./transaction.queries";

export const useGetQuote = () => {
  return useMutation({
    mutationFn: (payload: QuotePayload) => getQuote(payload),
  });
};

export const useExecuteConversion = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (quoteId: string) => executeConversion(quoteId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: walletKeys.all });
      qc.invalidateQueries({ queryKey: transactionKeys.all });
    },
  });
};
