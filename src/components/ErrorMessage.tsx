import type { AxiosError } from "axios";
import type { ApiError } from "../types";

interface Props {
  error: Error | AxiosError | null;
}

const ErrorMessage = ({ error }: Props) => {
  if (!error) return null;

  const axiosErr = error as AxiosError<ApiError>;
  const message =
    axiosErr.response?.data?.error ?? error.message ?? "Something went wrong";
  const details = axiosErr.response?.data?.details;

  return (
    <div className="rounded-md bg-red-50 border border-red-200 p-3">
      <p className="text-sm text-red-700 font-medium">{message}</p>
      {details && (
        <ul className="mt-1 list-disc list-inside">
          {Object.entries(details).map(([field, msg]) => (
            <li key={field} className="text-xs text-red-600">
              {field}: {msg}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};
export default ErrorMessage;
