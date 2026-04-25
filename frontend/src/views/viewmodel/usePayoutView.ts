import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreatePayout } from "../../queries/payout.queries";
import { generateIdempotencyKey } from "../../utils/idempotency";
import { useAuthStore } from "../../store/authStore";
import type { Currency } from "../../types";

export const PAYOUT_CURRENCIES: Currency[] = ["NGN", "KES"];

export const usePayoutView = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useCreatePayout();
  const userId = useAuthStore((s) => s.user?.id);

  const [sourceCurrency, setSourceCurrency] = useState<Currency>("NGN");
  const [amount, setAmount] = useState("");
  const [recipient, setRecipient] = useState({
    accountNumber: "",
    bankCode: "",
    accountName: "",
  });
  const [sessionId] = useState(generateIdempotencyKey);
  const idempotencyKey = useMemo(
    () =>
      `${recipient.accountNumber}-${userId}-${sourceCurrency}-${amount}-${sessionId}`,
    [recipient.accountNumber, userId, sourceCurrency, amount, sessionId],
  );

  const setRecipientField = (field: keyof typeof recipient) => {
    return (e: React.ChangeEvent<HTMLInputElement>) =>
      setRecipient((r) => ({ ...r, [field]: e.target.value }));
  };

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    mutate(
      { sourceCurrency, amount, recipient, idempotencyKey },
      {
        onSuccess: () => {
          toast.success("Payout submitted successfully");
          navigate("/transactions");
        },
      },
    );
  };

  return {
    sourceCurrency,
    setSourceCurrency,
    amount,
    setAmount,
    recipient,
    setRecipientField,
    isPending,
    error,
    handleSubmit,
  };
};
