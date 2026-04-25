import api from "./api";
import type { Currency, ConversionQuote, ConversionResponse } from "../types";

export interface QuotePayload {
  sourceCurrency: Currency;
  targetCurrency: Currency;
  amountIn: string;
}

export const getQuote = async (
  payload: QuotePayload,
): Promise<ConversionQuote> => {
  const refinedPayloadObj = {
    source_currency: payload?.sourceCurrency,
    target_currency: payload?.targetCurrency,
    source_amount: Number(payload?.amountIn) * 100,
  };
  const { data } = await api.post<ConversionQuote>(
    "/conversions/quote",
    refinedPayloadObj,
  );
  return data;
};

export const executeConversion = async (
  quoteId: string,
): Promise<ConversionResponse> => {
  const { data } = await api.post<ConversionResponse>("/conversions/execute", {
    quote_id: quoteId,
  });
  return data;
};
