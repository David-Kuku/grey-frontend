import api from "./api";
import type { Currency, DepositResponse } from "../types";

export interface DepositPayload {
  currency: Currency;
  amount: string;
  idempotencyKey: string;
}

export const createDeposit = async (
  payload: DepositPayload,
): Promise<DepositResponse> => {
  const refinedPayloadObj = {
    currency: payload?.currency,
    amount: Number(payload?.amount) * 100,
  };
  const { data } = await api.post<DepositResponse>("/deposits", refinedPayloadObj, {
    headers: { "idempotency-key": payload.idempotencyKey },
  });
  return data;
};
