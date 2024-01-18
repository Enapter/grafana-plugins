import { Notificator } from '../notifications/notificator';
import {
  EnapterApiError,
  MisconfiguredPanelError,
  UnknownCommandError,
} from '../hooks/useRunCommandWorkflow';
import { getErrorMessage } from './getErrorMessage';
import { GrafanaApiError } from '../api/client';

export const handleError = (e: unknown, description?: string) => {
  if (!e) {
    return;
  }

  console.error(e);

  const notificator = Notificator.getInstance();

  if (typeof e === 'object' && 'data' in e) {
    handleError(e.data, description);

    return;
  }

  if (e instanceof MisconfiguredPanelError) {
    notificator.warning(getErrorMessage(e.message), description);

    return;
  }

  if (e instanceof EnapterApiError) {
    notificator.error(getErrorMessage(e.message), description);

    return;
  }

  if (e instanceof UnknownCommandError) {
    notificator.error('Unknown command error');

    return;
  }

  if (e instanceof GrafanaApiError) {
    notificator.error(getErrorMessage(e.message), description);

    return;
  }

  notificator.error('Something went wrong');
};
