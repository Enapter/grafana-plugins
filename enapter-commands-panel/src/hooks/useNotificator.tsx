import { useMemo } from 'react';
import { Notificator } from '../notifications/notificator';

export const useNotificator = () => {
  return useMemo(() => Notificator.getInstance(), []);
};
