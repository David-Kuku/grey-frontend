import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createDeposit, type DepositPayload } from '../services/deposit.service'
import { walletKeys } from './wallet.queries'
import { transactionKeys } from './transaction.queries'

export function useDeposit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: DepositPayload) => createDeposit(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: walletKeys.all })
      qc.invalidateQueries({ queryKey: transactionKeys.all })
    },
  })
}
