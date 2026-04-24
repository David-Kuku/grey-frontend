import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDeposit } from "../../queries/deposit.queries";
import { generateIdempotencyKey } from "../../utils/idempotency";
import type { Currency } from "../../types";

export const useDepositView = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useDeposit();
  const [currency, setCurrency] = useState<Currency>("USD");
  const [amount, setAmount] = useState("");
  const [idempotencyKey] = useState(generateIdempotencyKey);

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    mutate(
      { currency, amount, idempotencyKey },
      { onSuccess: () => navigate("/") },
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
