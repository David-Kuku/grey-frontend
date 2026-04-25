import type { AxiosError } from "axios";
import type { ApiError } from "../types";

interface Props {
  error: Error | AxiosError | null;
}

const ErrorMessage = ({ error }: Props) => {
  if (!error) return null;

  const axiosErr = error as AxiosError<ApiError>;
  const message = axiosErr.response?.data?.message ?? "Something went wrong";

  return (
    <div className="rounded-md bg-red-50 border border-red-200 p-3">
      <p className="text-sm text-red-700 font-medium">{message}</p>
    </div>
  );
};
export default ErrorMessage;
