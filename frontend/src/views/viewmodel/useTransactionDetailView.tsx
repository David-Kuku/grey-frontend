import { useNavigate, useParams } from "react-router-dom";
import { useTransaction } from "../../queries/transaction.queries";
import { formatAmount } from "../../utils/currency";
import { formatDate } from "../../utils/date";
import StatusBadge from "../../components/StatusBadge";
import type { Currency } from "../../types";

export const useTransactionDetailView = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: txn, isLoading, error } = useTransaction(id!);

  const rows: [string, React.ReactNode][] = txn
    ? [
        ["ID", <span className="font-mono text-xs">{txn.id}</span>],
        ["Type", <span className="capitalize">{txn.type}</span>],
        ["Status", <StatusBadge status={txn.status} />],
        ["Amount", formatAmount(txn.amount, txn.currency as Currency)],
        ...(txn.targetCurrency
          ? [
              [
                "Converted to",
                formatAmount(txn.targetAmount!, txn.targetCurrency as Currency),
              ] as [string, React.ReactNode],
            ]
          : []),
        ["Currency", txn.currency],
        ["Date", formatDate(txn.createdAt)],
      ]
    : [];

  if (txn?.metadata) {
    const meta = txn.metadata as Record<string, string>;
    if (meta.accountName) rows.push(["Recipient", meta.accountName]);
    if (meta.accountNumber) rows.push(["Account", meta.accountNumber]);
    if (meta.bankCode) rows.push(["Bank code", meta.bankCode]);
  }

  return { txn, isLoading, error, navigate, rows };
};
