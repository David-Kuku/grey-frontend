import api from "./api";
import type { Currency, Payout, RecipientDetails } from "../types";

export interface PayoutPayload {
  sourceCurrency: Currency;
  amount: string;
  recipient: RecipientDetails;
}

export const createPayout = async (payload: PayoutPayload): Promise<Payout> => {
  const { data } = await api.post<Payout>("/payouts", payload);
  return data;
};

export const getPayout = async (id: string): Promise<Payout> => {
  const { data } = await api.get<Payout>(`/payouts/${id}`);
  return data;
};
