import { useTransactionDetailView } from "../viewmodel/useTransactionDetailView";

export default function TransactionDetailPage() {
  const { txn, isLoading, error, navigate, rows } = useTransactionDetailView();

  if (isLoading) {
    return (
      <div className="max-w-md">
        <div className="h-48 bg-gray-100 rounded-xl animate-pulse" />
      </div>
    );
  }

  if (error || !txn) {
    return (
      <div className="max-w-md">
        <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-sm text-red-700">
          Transaction not found.
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-md">
      <button
        onClick={() => navigate(-1)}
        className="text-sm text-gray-500 hover:text-gray-900 mb-4 flex items-center gap-1"
      >
        ← Back
      </button>

      <h2 className="text-xl font-semibold text-gray-900 mb-6 capitalize">
        {txn.transaction_type} detail
      </h2>

      <div className="bg-white rounded-xl border border-gray-200 divide-y divide-gray-100">
        {rows.map(([label, value]) => (
          <div
            key={label}
            className="flex items-start justify-between px-4 py-3"
          >
            <span className="text-sm text-gray-500 w-32 shrink-0">{label}</span>
            <span className="text-sm text-gray-900 text-right">{value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
