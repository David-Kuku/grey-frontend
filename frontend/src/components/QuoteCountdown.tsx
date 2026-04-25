import { useEffect, useState } from "react";
import { secondsUntil } from "../utils/date";

interface Props {
  expiresAt: string;
  onExpire: () => void;
}

const QuoteCountdown = ({ expiresAt, onExpire }: Props) => {
  const [seconds, setSeconds] = useState(() => secondsUntil(expiresAt));

  useEffect(() => {
    if (seconds <= 0) {
      onExpire();
      return;
    }
    const id = setInterval(() => {
      const s = secondsUntil(expiresAt);
      setSeconds(s);
      if (s <= 0) {
        clearInterval(id);
        onExpire();
      }
    }, 1000);
    return () => clearInterval(id);
  }, [expiresAt, onExpire, seconds]);

  const isUrgent = seconds <= 10;

  return (
    <span
      className={`font-mono text-sm font-medium ${isUrgent ? "text-red-600" : "text-gray-600"}`}
    >
      {seconds}s
    </span>
  );
};

export default QuoteCountdown;
