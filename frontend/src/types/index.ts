export type Currency = "USD" | "GBP" | "EUR" | "NGN" | "KES";

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
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
  id: string;
  balances: CurrencyBalance[];
}

export interface Deposit {
  id: string;
  currency: Currency;
  amount: string;
  status: "completed";
  createdAt: string;
  idempotencyKey: string;
}

export interface ConversionQuote {
  id: string;
  sourceCurrency: Currency;
  targetCurrency: Currency;
  amountIn: string;
  amountOut: string;
  rate: string;
  fee: string;
  expiresAt: string;
}

export interface Conversion {
  id: string;
  sourceCurrency: Currency;
  targetCurrency: Currency;
  amountIn: string;
  amountOut: string;
  rate: string;
  fee: string;
  status: "completed";
  createdAt: string;
}

export interface RecipientDetails {
  accountNumber: string;
  bankCode: string;
  accountName: string;
}

export type PayoutStatus = "pending" | "processing" | "successful" | "failed";

export interface Payout {
  id: string;
  sourceCurrency: Currency;
  amount: string;
  status: PayoutStatus;
  recipient: RecipientDetails;
  createdAt: string;
  updatedAt: string;
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
  transaction_type: TransactionType;
  status: TransactionStatus;
  amount: string;
  currency: Currency;
  created_at: string;
  metadata?: Record<string, unknown>;
  user_id: string;
}

export interface PaginatedTransactions {
  transactions: Transaction[];
  page: number;
  limit: number;
  total: number;
  hasMore: boolean;
}

export interface ApiError {
  message: string;
  code: number;
}
