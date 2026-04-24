import { useQuery } from "@tanstack/react-query";
import { getWallet } from "../services/wallet.service";

export const walletKeys = {
  all: ["wallet"] as const,
};

export const useWallet = () => {
  return useQuery({
    queryKey: walletKeys.all,
    queryFn: getWallet,
  });
};
