import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useGetQuote,
  useExecuteConversion,
} from "../../queries/conversion.queries";
import type { ConversionQuote, Currency } from "../../types";

type Step = "form" | "confirm";

export function useConvertView() {
  const navigate = useNavigate();
  const {
    mutate: getQuote,
    isPending: quoting,
    error: quoteError,
  } = useGetQuote();
  const {
    mutate: execute,
    isPending: executing,
    error: execError,
  } = useExecuteConversion();

  const [step, setStep] = useState<Step>("form");
  const [quote, setQuote] = useState<ConversionQuote | null>(null);
  const [expired, setExpired] = useState(false);

  const [sourceCurrency, setSourceCurrency] = useState<Currency>("USD");
  const [targetCurrency, setTargetCurrency] = useState<Currency>("NGN");

  function handleSourceCurrencyUpdate(currency: Currency) {
    setSourceCurrency(currency);
    if (currency === targetCurrency) {
      setTargetCurrency(sourceCurrency);
    }
  }
  const [amountIn, setAmountIn] = useState("");

  function handleGetQuote(e: React.FormEvent) {
    e.preventDefault();
    setExpired(false);
    getQuote(
      { sourceCurrency, targetCurrency, amountIn },
      {
        onSuccess: (q) => {
          setQuote(q);
          setStep("confirm");
        },
      },
    );
  }

  function handleExecute() {
    if (!quote) return;
    execute(quote.quote_id, {
      onSuccess: () => {
        toast.success("Conversion successful");
        navigate("/");
      },
    });
  }

  const handleExpire = useCallback(() => setExpired(true), []);

  return {
    step,
    setStep,
    quote,
    expired,
    sourceCurrency,
    setSourceCurrency: handleSourceCurrencyUpdate,
    targetCurrency,
    setTargetCurrency,
    amountIn,
    setAmountIn,
    quoting,
    quoteError,
    executing,
    execError,
    handleGetQuote,
    handleExecute,
    handleExpire,
  };
}
