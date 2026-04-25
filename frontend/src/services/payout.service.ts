import api from "./api";
import type { Currency, PayoutResponse, RecipientDetails } from "../types";

export interface PayoutPayload {
  sourceCurrency: Currency;
  amount: string;
  recipient: RecipientDetails;
  idempotencyKey: string;
}

export const createPayout = async (payload: PayoutPayload): Promise<PayoutResponse> => {
  const refinedPayloadObj = {
    source_currency: payload?.sourceCurrency,
    amount: Number(payload?.amount) * 100,
    recipient_name: payload?.recipient?.accountName,
    recipient_bank_code: payload?.recipient?.bankCode,
    recipient_account: payload?.recipient?.accountNumber,
  };

  const { data } = await api.post<PayoutResponse>("/payouts", refinedPayloadObj, {
    headers: { "idempotency-key": payload.idempotencyKey },
  });
  return data;
};
