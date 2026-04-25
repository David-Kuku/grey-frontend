import api from "./api";
import type { Currency, Payout, RecipientDetails } from "../types";

export interface PayoutPayload {
  sourceCurrency: Currency;
  amount: string;
  recipient: RecipientDetails;
  idempotencyKey: string;
}

export const createPayout = async (payload: PayoutPayload): Promise<Payout> => {
  const refinedPayloadObj = {
    source_currency: payload?.sourceCurrency,
    amount: Number(payload?.amount) * 100,
    recipient_name: payload?.recipient?.accountName,
    recipient_bank_code: payload?.recipient?.bankCode,
    recipient_account: payload?.recipient?.accountNumber,
  };

  const { data } = await api.post<Payout>("/payouts", refinedPayloadObj, {
    headers: { "idempotency-key": payload.idempotencyKey },
  });
  return data;
};

export const getPayout = async (id: string): Promise<Payout> => {
  const { data } = await api.get<Payout>(`/payouts/${id}`);
  return data;
};
