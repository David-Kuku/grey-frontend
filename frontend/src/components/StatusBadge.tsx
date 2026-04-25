import type { TransactionStatus } from "../types";

const STYLES: Record<TransactionStatus, string> = {
  completed: "bg-green-100 text-green-700",
  successful: "bg-green-100 text-green-700",
  pending: "bg-yellow-100 text-yellow-700",
  processing: "bg-blue-100 text-blue-700",
  failed: "bg-red-100 text-red-700",
};

interface Props {
  status: TransactionStatus;
}

const StatusBadge = ({ status }: Props) => {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${STYLES[status]}`}
    >
      {status}
    </span>
  );
};

export default StatusBadge;
