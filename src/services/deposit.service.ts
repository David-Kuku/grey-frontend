import api from "./api";
import type { Currency, Deposit } from "../types";

export interface DepositPayload {
  currency: Currency;
  amount: string;
  idempotencyKey: string;
}

export const createDeposit = async (
  payload: DepositPayload,
): Promise<Deposit> => {
  const { data } = await api.post<Deposit>("/deposits", payload, {
    headers: { "Idempotency-Key": payload.idempotencyKey },
  });
  return data;
};
