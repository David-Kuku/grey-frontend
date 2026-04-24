import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useCreatePayout } from "../../queries/payout.queries";
import type { Currency } from "../../types";

export const PAYOUT_CURRENCIES: Currency[] = ["NGN", "KES"];

export const usePayoutView = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useCreatePayout();

  const [sourceCurrency, setSourceCurrency] = useState<Currency>("NGN");
  const [amount, setAmount] = useState("");
  const [recipient, setRecipient] = useState({
    accountNumber: "",
    bankCode: "",
    accountName: "",
  });

  const setRecipientField = (field: keyof typeof recipient) => {
    return (e: React.ChangeEvent<HTMLInputElement>) =>
      setRecipient((r) => ({ ...r, [field]: e.target.value }));
  };

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    mutate(
      { sourceCurrency, amount, recipient },
      { onSuccess: (p) => navigate(`/transactions/${p.id}`) },
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
