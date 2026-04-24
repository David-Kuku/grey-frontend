import { usePayoutView, PAYOUT_CURRENCIES } from "../viewmodel/usePayoutView";
import ErrorMessage from "../../components/ErrorMessage";
import type { Currency } from "../../types";

export default function PayoutPage() {
  const {
    sourceCurrency,
    setSourceCurrency,
    amount,
    setAmount,
    recipient,
    setRecipientField,
    isPending,
    error,
    handleSubmit,
  } = usePayoutView();

  return (
    <div className="max-w-md">
      <h2 className="text-xl font-semibold text-gray-900 mb-1">Send payout</h2>
      <p className="text-sm text-gray-500 mb-6">
        Send funds to a local bank account in NGN or KES.
      </p>

      <div className="bg-white rounded-xl border border-gray-200 p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Source currency
              </label>
              <select
                value={sourceCurrency}
                onChange={(e) => setSourceCurrency(e.target.value as Currency)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 bg-white"
              >
                {PAYOUT_CURRENCIES.map((c) => (
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
          </div>

          <hr className="border-gray-100" />
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">
            Recipient details
          </p>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Account name
            </label>
            <input
              type="text"
              required
              value={recipient.accountName}
              onChange={setRecipientField("accountName")}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              placeholder="e.g. John Doe"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Account number
            </label>
            <input
              type="text"
              required
              value={recipient.accountNumber}
              onChange={setRecipientField("accountNumber")}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              placeholder="10-digit account number"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Bank code
            </label>
            <input
              type="text"
              required
              value={recipient.bankCode}
              onChange={setRecipientField("bankCode")}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
              placeholder="e.g. 058"
            />
          </div>

          {error && <ErrorMessage error={error} />}

          <button
            type="submit"
            disabled={isPending}
            className="w-full bg-gray-900 text-white rounded-lg py-2 text-sm font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isPending ? "Sending…" : "Send payout"}
          </button>
        </form>
      </div>
    </div>
  );
}
