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

  const meta = txn?.metadata as Record<string, unknown> | undefined;
  const targetCurrency = meta?.target_currency as Currency | undefined;
  const targetAmount = meta?.target_amount as number | undefined;

  const rows: [string, React.ReactNode][] = txn
    ? [
        ["ID", <span className="font-mono text-xs">{txn.id}</span>],
        ["Type", <span className="capitalize">{txn.transaction_type}</span>],
        ["Status", <StatusBadge status={txn.status} />],
        ["Amount", formatAmount(txn.amount / 100, txn.currency as Currency)],
        ...(targetCurrency && targetAmount !== undefined
          ? [
              [
                "Converted to",
                formatAmount(targetAmount / 100, targetCurrency),
              ] as [string, React.ReactNode],
            ]
          : []),
        ["Currency", txn.currency],
        ["Date", formatDate(txn.created_at)],
      ]
    : [];

  if (meta) {
    const m = meta as Record<string, string>;
    if (m.recipient_name) rows.push(["Recipient", m.recipient_name]);
    if (m.recipient_account) rows.push(["Account", m.recipient_account]);
    if (m.recipient_bank) rows.push(["Bank code", m.recipient_bank]);
  }

  return { txn, isLoading, error, navigate, rows };
};
