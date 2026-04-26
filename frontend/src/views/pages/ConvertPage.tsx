import { useConvertView } from "../viewmodel/useConvertView";
import { SUPPORTED_CURRENCIES, formatAmount } from "../../utils/currency";
import { formatDate } from "../../utils/date";
import ErrorMessage from "../../components/ErrorMessage";
import QuoteCountdown from "../../components/QuoteCountdown";
import type { Currency } from "../../types";

export default function ConvertPage() {
  const {
    step,
    setStep,
    quote,
    expired,
    sourceCurrency,
    setSourceCurrency,
    targetCurrency,
    setTargetCurrency,
    amountIn,
    setAmountIn,
    quoting,
    quoteError,
    executing,
    execError,
    handleGetQuote,
    handleExecute,
    handleExpire,
  } = useConvertView();

  return (
    <div className="max-w-md">
      <h2 className="text-xl font-semibold text-gray-900 mb-1">
        Convert currency
      </h2>
      <p className="text-sm text-gray-500 mb-6">
        Get a live quote and confirm to convert.
      </p>

      {step === "form" && (
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <form onSubmit={handleGetQuote} className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  From
                </label>
                <select
                  value={sourceCurrency}
                  onChange={(e) => {
                    setSourceCurrency(e.target.value as Currency);
                  }}
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
                  To
                </label>
                <select
                  value={targetCurrency}
                  onChange={(e) =>
                    setTargetCurrency(e.target.value as Currency)
                  }
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 bg-white"
                >
                  {SUPPORTED_CURRENCIES.filter((c) => c !== sourceCurrency).map(
                    (c) => (
                      <option key={c} value={c}>
                        {c}
                      </option>
                    ),
                  )}
                </select>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Amount ({sourceCurrency})
              </label>
              <input
                type="number"
                required
                min="0.01"
                step="0.01"
                value={amountIn}
                onChange={(e) => setAmountIn(e.target.value)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
                placeholder="0.00"
              />
            </div>

            {quoteError && <ErrorMessage error={quoteError} />}

            <button
              type="submit"
              disabled={quoting}
              className="w-full bg-gray-900 text-white rounded-lg py-2 text-sm font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {quoting ? "Getting quote…" : "Get quote"}
            </button>
          </form>
        </div>
      )}

      {step === "confirm" && quote && (
        <div className="bg-white rounded-xl border border-gray-200 p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-900">Quote details</h3>
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-400">Expires in</span>
              {expired ? (
                <span className="text-xs font-medium text-red-600">
                  Expired
                </span>
              ) : (
                <QuoteCountdown
                  expiresAt={quote.expires_at}
                  onExpire={handleExpire}
                />
              )}
            </div>
          </div>

          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-500">You send</span>
              <span className="font-medium">
                {formatAmount(quote.source_amount / 100, quote.source_currency)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">You receive</span>
              <span className="font-medium text-green-700">
                {formatAmount(quote.target_amount / 100, quote.target_currency)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Rate</span>
              <span className="text-gray-700">
                1 {quote.source_currency} ={" "}
                {parseFloat(quote.quoted_rate).toFixed(4)}{" "}
                {quote.target_currency}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Fee</span>
              <span className="text-gray-700">
                {formatAmount(quote.fee_amount / 100, quote.fee_currency)}
              </span>
            </div>
            <div className="flex justify-between text-xs text-gray-400">
              <span>Quote ID</span>
              <span className="font-mono">{quote.quote_id.slice(0, 8)}…</span>
            </div>
            <div className="flex justify-between text-xs text-gray-400">
              <span>Expires</span>
              <span>{formatDate(quote.expires_at)}</span>
            </div>
          </div>

          {expired && (
            <div className="rounded-md bg-yellow-50 border border-yellow-200 p-3">
              <p className="text-sm text-yellow-700">
                This quote has expired. Please get a new quote.
              </p>
            </div>
          )}

          {execError && <ErrorMessage error={execError} />}

          <div className="flex gap-3">
            <button
              onClick={() => setStep("form")}
              className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm font-medium hover:bg-gray-50 transition-colors"
            >
              Back
            </button>
            <button
              onClick={handleExecute}
              disabled={executing || expired}
              className="flex-1 bg-gray-900 text-white rounded-lg py-2 text-sm font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {executing ? "Converting…" : "Confirm"}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
