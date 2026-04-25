import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useDeposit } from "../../queries/deposit.queries";
import { generateIdempotencyKey } from "../../utils/idempotency";
import { useAuthStore } from "../../store/authStore";
import type { Currency } from "../../types";

export const useDepositView = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useDeposit();
  const userId = useAuthStore((s) => s.user?.id);
  const [currency, setCurrency] = useState<Currency>("USD");
  const [amount, setAmount] = useState("");
  const [sessionId] = useState(generateIdempotencyKey);
  const idempotencyKey = useMemo(
    () => `${userId}-${currency}-${amount}-${sessionId}`,
    [userId, currency, amount, sessionId],
  );

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    mutate(
      { currency, amount, idempotencyKey },
      {
        onSuccess: () => {
          toast.success("Deposit successful");
          navigate("/transactions");
        },
      },
    );
  };

  return {
    currency,
    setCurrency,
    amount,
    setAmount,
    isPending,
    error,
    handleSubmit,
  };
};
