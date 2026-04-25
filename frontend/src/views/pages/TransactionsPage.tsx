import { Link } from "react-router-dom";
import {
  useTransactionsView,
  PAGE_SIZE_OPTIONS,
} from "../viewmodel/useTransactionsView";
import { formatAmount } from "../../utils/currency";
import { formatDate } from "../../utils/date";
import StatusBadge from "../../components/StatusBadge";
import type { Currency } from "../../types";

export default function TransactionsPage() {
  const {
    data,
    isLoading,
    isFetching,
    page,
    setPage,
    limit,
    handleLimitChange,
  } = useTransactionsView();

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-900">Transactions</h2>
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span>Rows per page</span>
          <select
            value={limit}
            onChange={(e) => handleLimitChange(Number(e.target.value))}
            className="border border-gray-200 rounded-lg px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 bg-white"
          >
            {PAGE_SIZE_OPTIONS.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <div
              key={i}
              className="h-14 bg-gray-100 rounded-lg animate-pulse"
            />
          ))}
        </div>
      ) : data?.transactions?.length === 0 ? (
        <div className="text-center py-16 text-gray-400 text-sm">
          No transactions found.
        </div>
      ) : (
        <div
          className={`bg-white rounded-xl border border-gray-200 divide-y divide-gray-100 transition-opacity ${isFetching ? "opacity-50" : "opacity-100"}`}
        >
          {data?.transactions?.map((txn) => (
            <Link
              key={txn.id}
              to={`/transactions/${txn.id}`}
              className="flex items-center justify-between px-4 py-3 hover:bg-gray-50 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center">
                  <span className="text-xs font-medium">
                    {txn?.transaction_type?.[0]?.toUpperCase()}
                  </span>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900 capitalize">
                    {txn?.transaction_type}
                  </p>
                  <p className="text-xs text-gray-400">
                    {formatDate(txn?.created_at)}
                  </p>
                </div>
              </div>
              <div className="text-right space-y-0.5">
                <p className="text-sm font-medium text-gray-900">
                  {formatAmount(
                    Number(txn.amount) / 100,
                    txn.currency as Currency,
                  )}
                </p>
                <StatusBadge status={txn.status} />
              </div>
            </Link>
          ))}
        </div>
      )}

      {data && (data.has_more || page > 1) && (
        <div className="flex items-center justify-between mt-4">
          <button
            disabled={page === 1}
            onClick={() => setPage((p) => p - 1)}
            className="text-sm text-gray-500 hover:text-gray-900 disabled:opacity-30 disabled:cursor-not-allowed"
          >
            ← Previous
          </button>
          <span className="text-xs text-gray-400">Page {page}</span>
          <button
            disabled={!data.has_more}
            onClick={() => setPage((p) => p + 1)}
            className="text-sm text-gray-500 hover:text-gray-900 disabled:opacity-30 disabled:cursor-not-allowed"
          >
            Next →
          </button>
        </div>
      )}
    </div>
  );
}
