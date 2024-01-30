import { IconName, SelectableValue } from '@grafana/data';
import { BlueprintManifest, Command as ManifestCommand } from 'types/manifest-schema';

export type ArgumentOption = Pick<SelectableValue<string>, 'value' | 'label' | 'description'>;

export enum OriginType {
  Populate = 'populate',
  Fixed = 'fixed',
}

export interface BasicArgument {
  key: string;
  displayName: string;
  description?: string;
  required?: boolean;
  isValid: boolean;
  errorMessage?: string;
  originType: OriginType;
}

export interface IntegerArgument extends BasicArgument {
  type: 'integer';
  value?: string;
  defaultValue?: number;
  min?: number;
  max?: number;
  options: ArgumentOption[];
}

export interface FloatArgument extends BasicArgument {
  type: 'float';
  value?: string;
  defaultValue?: number;
  min?: number;
  max?: number;
  options: ArgumentOption[];
}

export interface StringArgument extends BasicArgument {
  type: 'string';
  value?: string;
  defaultValue?: string;
  options: ArgumentOption[];
}

export interface BooleanArgument extends BasicArgument {
  type: 'boolean';
  value?: boolean;
  defaultValue?: boolean;
}

export type Argument = IntegerArgument | FloatArgument | StringArgument | BooleanArgument;

export enum ConfirmationType {
  Always = 'always',
  Invalid = 'invalid',
  Never = 'never',
}

export type Command = {
  key: string;
  displayName: string;
  description?: string;
  arguments?: Record<string, Argument>;
  populateValuesCommand?: string;
  confirmationType: ConfirmationType;
} & Pick<ManifestCommand, 'confirmation'>;

export type PanelState = {
  pluginVersion: '1.0';

  datasource?: {
    name: string;
    uid: string;
  };

  deviceId: string;

  currentCommand?: Command;
  commands: Record<string, Command>;
  manifestCommands: NonNullable<BlueprintManifest['commands']>;

  appearance: {
    icon: IconName;
    buttonText?: string;
    bgColor: string;
    textColor: string;
    fullWidth: boolean;
    fullHeight: boolean;
    shouldScaleText: boolean;
  };
};
