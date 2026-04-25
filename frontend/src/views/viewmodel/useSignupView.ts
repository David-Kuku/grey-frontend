import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSignup } from "../../queries/auth.queries";

export const useSignupView = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useSignup();
  const [form, setForm] = useState({
    email: "",
    password: "",
    firstName: "",
    lastName: "",
  });

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    mutate(form, { onSuccess: () => navigate("/") });
  };

  const set = (field: keyof typeof form) => {
    return (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));
  };

  return { form, set, isPending, error, handleSubmit };
};
