import { useDepositView } from "../viewmodel/useDepositView";
import { SUPPORTED_CURRENCIES } from "../../utils/currency";
import ErrorMessage from "../../components/ErrorMessage";
import type { Currency } from "../../types";

export default function DepositPage() {
  const {
    currency,
    setCurrency,
    amount,
    setAmount,
    isPending,
    error,
    handleSubmit,
  } = useDepositView();

  return (
    <div className="max-w-md">
      <h2 className="text-xl font-semibold text-gray-900 mb-1">
        Deposit funds
      </h2>
      <p className="text-sm text-gray-500 mb-6">
        Add balance to one of your currency wallets.
      </p>

      <div className="bg-white rounded-xl border border-gray-200 p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Currency
            </label>
            <select
              value={currency}
              onChange={(e) => setCurrency(e.target.value as Currency)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 bg-white"
            >
              {SUPPORTED_CURRENCIES.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Amount
            </label>
            <input
              type="number"
              required
              min="0.01"
              step="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              placeholder="0.00"
            />
          </div>

          {error && <ErrorMessage error={error} />}

          <button
            type="submit"
            disabled={isPending}
            className="w-full bg-gray-900 text-white rounded-lg py-2 text-sm font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isPending ? "Processing…" : "Deposit"}
          </button>
        </form>
      </div>
    </div>
  );
}
