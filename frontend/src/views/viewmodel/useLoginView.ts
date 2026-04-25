import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useLogin } from "../../queries/auth.queries";

export const useLoginView = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useLogin();
  const [form, setForm] = useState({ email: "", password: "" });

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    mutate(form, { onSuccess: () => navigate("/") });
  };

  return { form, setForm, isPending, error, handleSubmit };
};
