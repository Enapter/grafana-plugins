import { useState } from 'react';
import { apiClient } from '../components/Editor';
import { Argument, Command, PanelState } from '../types/types';
import { mergeArgumentsWithPopulatedValues } from '../utils/mergeArgumentsWithPrepopulatedValues';
import { stringify } from '../utils/stringify';
import { ArgValidator } from '../validation/arg-validator';
import { replaceVariables } from '../utils/replaceVariables';

export class MisconfiguredPanelError extends Error {}

export class EnapterApiError extends Error {}

export class UnknownCommandError extends Error {}

const getArgValue = (arg: Argument) => {
  if (arg.type === 'boolean') {
    return Boolean(arg.value);
  }

  if (arg.value === null || arg.value === undefined) {
    return null;
  }

  if (arg.type === 'float') {
    return Number.parseFloat(replaceVariables(arg.value));
  }

  if (arg.type === 'integer') {
    return Number.parseInt(replaceVariables(arg.value), 10);
  }

  return replaceVariables(arg.value);
};

const getArgsForRequest = (args: Command['arguments']) => {
  const argsForRequest: Record<string, any> = {};

  if (!args) {
    return argsForRequest;
  }

  for (const [argName, arg] of Object.entries(args)) {
    const value = getArgValue(arg);

    if (value === null || Number.isNaN(value)) {
      continue;
    }

    argsForRequest[argName] = value;
  }

  return argsForRequest;
};

const checkAnyArgInvalid = (args: Record<string, Argument>) => {
  return Object.values(args)
    .map((arg) => {
      return new ArgValidator(arg).isValueValid();
    })
    .some((isValid) => !isValid);
};

const handleRunCommandResult = ({ state, errors, payload }: any) => {
  if (errors.length) {
    throw new EnapterApiError(
      errors
        .map((e: any) => {
          if (e.payload?.code || e.payload?.reason) {
            return [e.payload.code, e.payload.reason].filter(Boolean).join(': ');
          }

          if (e.code || e.reason || e.message) {
            return [e.code, e.reason || e.message].filter(Boolean).join(': ');
          }

          if (e.message) {
            return e.message;
          }

          return stringify(e);
        })
        .join(', ')
    );
  }

  if (state === 'succeeded') {
    return payload;
  }

  throw new UnknownCommandError();
};

export const useRunCommandWorkflow = ({
  deviceId,
  datasource,
}: {
  deviceId: PanelState['deviceId'];
  datasource: PanelState['datasource'];
}) => {
  const [isRunning, setIsRunning] = useState(false);
  const [isPopulating, setIsPopulating] = useState(false);

  const run = async (commandKey: string, args?: Command['arguments']) => {
    if (!datasource) {
      throw new MisconfiguredPanelError('Datasource is not selected');
    }

    try {
      setIsRunning(true);

      const argsForRequest = getArgsForRequest(args);

      return await apiClient.runCommand({
        queryType: 'command',
        datasource: {
          uid: datasource.uid,
        },
        payload: {
          commandName: commandKey,
          commandArgs: argsForRequest,
          deviceId,
        },
      });
    } catch (e) {
      throw e;
    } finally {
      setIsRunning(false);
    }
  };

  const populateValues = async (commandKey: string) => {
    try {
      setIsPopulating(true);
      const result = await run(commandKey);

      return handleRunCommandResult(result);
    } catch (e) {
      throw e;
    } finally {
      setIsPopulating(false);
    }
  };

  const runCommand = async (commandKey: string, args: Command['arguments']) => {
    if (args && checkAnyArgInvalid(args)) {
      throw new MisconfiguredPanelError('Some arguments are invalid');
    }

    const result = await run(commandKey, args);

    return handleRunCommandResult(result);
  };

  const populateValuesAndRunCommand = async (
    command: NonNullable<PanelState['currentCommand']>,
    args: Command['arguments']
  ) => {
    let commandArgs = args;

    if (command.populateValuesCommand) {
      const populatedValues = await populateValues(command.populateValuesCommand);
      commandArgs = mergeArgumentsWithPopulatedValues(args, populatedValues);
    }

    return runCommand(command.key, commandArgs);
  };

  return { isPopulating, populateValues, runCommand, isRunning, populateValuesAndRunCommand };
};
