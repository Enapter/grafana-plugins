import { stringify } from './stringify';
import { addDot } from './addDot';
import { capitalize } from './capitalize';

export const getErrorMessage = (...messages: any[]) => {
  return (messages.filter(Boolean) as string[])
    .map(stringify)
    .map(addDot)
    .map(capitalize)
    .join(' ');
};
