import { Link } from "react-router-dom";
import { useDashboardView } from "../viewmodel/useDashboardView";
import { formatAmount } from "../../utils/currency";
import { formatDate } from "../../utils/date";
import StatusBadge from "../../components/StatusBadge";
import type { Currency } from "../../types";

export default function DashboardPage() {
  const { wallet, walletLoading, txns, txnsLoading } = useDashboardView();

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Balances</h2>
        {walletLoading ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="h-24 bg-gray-100 rounded-xl animate-pulse"
              />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {wallet?.balances.map((b) => (
              <div
                key={b.currency}
                className="bg-white rounded-xl border border-gray-200 p-4"
              >
                <p className="text-xs font-medium text-gray-400 mb-1">
                  {b.currency}
                </p>
                <p className="text-lg font-semibold text-gray-900">
                  {formatAmount(b.amount, b.currency as Currency)}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>

      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900">
            Recent transactions
          </h2>
          <Link
            to="/transactions"
            className="text-sm text-gray-500 hover:text-gray-900"
          >
            View all
          </Link>
        </div>

        {txnsLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <div
                key={i}
                className="h-14 bg-gray-100 rounded-lg animate-pulse"
              />
            ))}
          </div>
        ) : txns?.data.length === 0 ? (
          <div className="text-center py-12 text-gray-400 text-sm">
            No transactions yet.{" "}
            <Link to="/deposit" className="text-gray-900 underline">
              Make a deposit
            </Link>{" "}
            to get started.
          </div>
        ) : (
          <div className="bg-white rounded-xl border border-gray-200 divide-y divide-gray-100">
            {txns?.data.map((txn) => (
              <Link
                key={txn.id}
                to={`/transactions/${txn.id}`}
                className="flex items-center justify-between px-4 py-3 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center">
                    <span className="text-xs">{txn.type[0].toUpperCase()}</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900 capitalize">
                      {txn.type}
                    </p>
                    <p className="text-xs text-gray-400">
                      {formatDate(txn.createdAt)}
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-900">
                    {formatAmount(txn.amount, txn.currency as Currency)}
                  </p>
                  <StatusBadge status={txn.status} />
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
