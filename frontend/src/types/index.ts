export type Currency = "USD" | "GBP" | "EUR" | "NGN" | "KES";

export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface CurrencyBalance {
  currency: Currency;
  balance_minor: number;
  balance: string;
}

export interface Wallet {
  balances: CurrencyBalance[];
}

export type TransactionType = "deposit" | "conversion" | "payout";
export type TransactionStatus =
  | "completed"
  | "pending"
  | "processing"
  | "successful"
  | "failed";

export interface Transaction {
  id: string;
  user_id: string;
  transaction_type: TransactionType;
  status: TransactionStatus;
  currency: Currency;
  amount: number;
  idempotency_key?: string;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface TransactionSummary extends Transaction {
  amount_formatted: string;
}

export interface PaginatedTransactions {
  transactions: TransactionSummary[];
  page: number;
  page_size: number;
  total_count: number;
  has_more: boolean;
}

export interface DepositResponse {
  transaction: Transaction;
}

export interface ConversionQuote {
  quote_id: string;
  source_currency: Currency;
  target_currency: Currency;
  source_amount: number;
  target_amount: number;
  market_rate: string;
  quoted_rate: string;
  spread_pct: string;
  fee_amount: number;
  fee_currency: Currency;
  expires_at: string;
  expires_in_seconds: number;
}

export interface ConversionResponse {
  transaction: Transaction;
  source_balance: number;
  target_balance: number;
  source_currency: Currency;
  target_currency: Currency;
}

export type PayoutStatus = "pending" | "processing" | "successful" | "failed";

export interface Payout {
  id: string;
  transaction_id: string;
  user_id: string;
  source_currency: Currency;
  amount: number;
  recipient_name: string;
  recipient_bank_code: string;
  recipient_account: string;
  status: PayoutStatus;
  failure_reason?: string;
  reversal_transaction_id?: string;
  created_at: string;
  updated_at: string;
}

export interface PayoutResponse {
  payout: Payout;
  transaction: Transaction;
  new_balance: number;
}

export interface RecipientDetails {
  accountNumber: string;
  bankCode: string;
  accountName: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, string>;
}
