import { PanelProps } from '@grafana/data';
import { Argument, PanelState } from '../types/types';
import { useImmer } from 'use-immer';
import { useCallback, useEffect, useMemo, useRef } from 'react';
import { current, Draft, produce } from 'immer';
import { replaceVariables } from '../utils/replaceVariables';
import { ArgValidator } from '../validation/arg-validator';

const getArgs = (props: PanelProps<{ commandButton: PanelState }>) => {
  const args: any = {};

  if (props.options.commandButton.currentCommand) {
    const command = props.options.commandButton.currentCommand;

    if (command.arguments) {
      for (const [argName, arg] of Object.entries(command.arguments)) {
        args[argName] = {
          ...arg,
          value: typeof arg.value === 'string' ? replaceVariables(arg.value) : arg.value,
        };
      }
    }
  }

  return args as Record<string, Argument>;
};

export const checkAnyArgInvalid = (args: Record<string, Argument>) => {
  return Object.values(args).some((arg) => !arg.isValid);
};

export const useNotPersistedArgs = (props: PanelProps<{ commandButton: PanelState }>) => {
  const [args, setArgs] = useImmer(getArgs(props));
  const isReset = useRef(false);

  useEffect(() => {
    setArgs(getArgs(props));
  }, [props, setArgs]);

  const updateArg = (argKey: string, recipe: (draft: Draft<Argument>) => void) => {
    if (!args[argKey]) {
      return;
    }

    isReset.current = false;

    setArgs((d) => {
      d[argKey] = produce(d[argKey], recipe);
    });
  };

  const hasInvalidArgs = useMemo(() => {
    return checkAnyArgInvalid(args);
  }, [args]);

  const resetArgs = useCallback(() => {
    if (isReset.current) {
      return;
    }

    setArgs(getArgs(props));
    isReset.current = true;
  }, [props, setArgs]);

  const validateArgs = () => {
    for (const argName of Object.keys(args)) {
      updateArg(argName, (draft) => {
        const argValidator = new ArgValidator(current(draft));

        draft.isValid = argValidator.isValueValid();
        draft.errorMessage = argValidator.getErrorMessage();
      });
    }
  };

  return { args, updateArg, updateArgs: setArgs, hasInvalidArgs, resetArgs, validateArgs };
};
