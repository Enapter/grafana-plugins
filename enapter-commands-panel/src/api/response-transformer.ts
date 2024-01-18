import { BlueprintManifest, CommandArgument } from 'types/manifest-schema';
import {
  Argument,
  BasicArgument,
  BooleanArgument,
  Command,
  ConfirmationType,
  FloatArgument,
  IntegerArgument,
  OriginType,
  StringArgument,
} from 'types/types';
import _toString from 'lodash/toString';
import { ArgValidator } from '../validation/arg-validator';

const getArgumentOptions = <T extends IntegerArgument | FloatArgument | StringArgument>(
  arg: CommandArgument
): T['options'] => {
  if (arg.type === 'integer' || arg.type === 'float' || arg.type === 'string') {
    const basicEnumOptions =
      arg.enum?.map((e) => ({ label: e.toString(), value: _toString(e) })) ?? [];

    const withMetaEnumOptions =
      arg.enum_with_metainfo?.map((e) => ({
        label: e.display_name,
        value: _toString(e.value),
        description: e.description,
      })) ?? [];

    return [...basicEnumOptions, ...withMetaEnumOptions].filter(
      (o) => o.value !== undefined
    ) as T['options'];
  }

  return [];
};

const getTypeDependentProperties = (arg: CommandArgument) => {
  if (arg.type === 'integer') {
    return {
      type: 'integer',
      value: undefined,
      min: arg.min,
      max: arg.max,
      options: getArgumentOptions<IntegerArgument>(arg),
    } as Omit<IntegerArgument, keyof BasicArgument>;
  }

  if (arg.type === 'float') {
    return {
      type: 'float',
      value: undefined,
      min: arg.min,
      max: arg.max,
      options: getArgumentOptions<FloatArgument>(arg),
    } as Omit<FloatArgument, keyof BasicArgument>;
  }

  if (arg.type === 'string') {
    return {
      type: 'string',
      value: undefined,
      options: getArgumentOptions<StringArgument>(arg),
    } as Omit<StringArgument, keyof BasicArgument>;
  }

  if (arg.type === 'boolean') {
    return {
      type: 'boolean',
      value: undefined,
    } as Omit<BooleanArgument, keyof BasicArgument>;
  }

  throw new Error(`Unknown argument type: ${arg.type}`);
};

const transformManifestArgumentToPanelArgument = (
  args: Record<string, CommandArgument>
): Record<string, Argument> => {
  const draft: Record<string, any> = {};

  Object.entries(args).forEach(([argName, arg]) => {
    const value = _toString(arg.default);
    const defaultValue = _toString(arg.default ?? undefined);

    const obj: any = {
      key: argName,
      displayName: arg.display_name,
      description: arg.description,
      required: arg.required,
      errorMessage: undefined,
      isValid: true,
      originType: OriginType.Populate,
      ...getTypeDependentProperties(arg),
      value,
      defaultValue,
    };

    const argValidator = new ArgValidator(obj);

    if (!argValidator.isValueValid()) {
      obj.isValid = false;
      obj.errorMessage = argValidator.getErrorMessage();
    }

    draft[argName] = obj;
  });

  return draft as Record<string, Argument>;
};

export const transformManifestCommandsToPanelCommands = (
  commands: BlueprintManifest['commands']
) => {
  const transformed: Record<string, Command> = {};

  if (!commands) {
    return transformed;
  }

  Object.entries(commands).forEach(([commandName, command]) => {
    const args: Record<string, Argument> | undefined = undefined;

    const draft: Command = {
      key: commandName,
      displayName: command.display_name,
      description: command.description,
      arguments: args,
      populateValuesCommand: command.populate_values_command,
      confirmation: command.confirmation,
      confirmationType:
        command.confirmation || command.populate_values_command
          ? ConfirmationType.Always
          : ConfirmationType.Never,
    };

    if (!command.arguments) {
      transformed[commandName] = draft;

      return;
    }

    draft.arguments = transformManifestArgumentToPanelArgument(command.arguments);

    transformed[commandName] = draft;
  });

  return transformed as Record<string, Command>;
};
