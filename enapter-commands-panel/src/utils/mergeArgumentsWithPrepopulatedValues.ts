import { Command, OriginType } from '../types/types';
import { current, produce } from 'immer';
import { ArgValidator } from '../validation/arg-validator';
import { stringify } from './stringify';

export const mergeArgumentsWithPopulatedValues = (
  args: Command['arguments'],
  populatedValues: any
) => {
  if (!args || !populatedValues || !Object.keys(populatedValues).length) {
    return args;
  }

  return produce(args, (draft) => {
    for (const [argName, arg] of Object.entries(current(draft))) {
      if (arg.originType !== OriginType.Populate) {
        continue;
      }

      if (argName in populatedValues) {
        arg.value = stringify(populatedValues[argName]);
        const argValidator = new ArgValidator(arg);
        arg.isValid = argValidator.isValueValid();
        arg.errorMessage = argValidator.getErrorMessage();
      }
    }
  });
};
