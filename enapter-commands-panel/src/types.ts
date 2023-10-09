import { ButtonProps } from '@grafana/ui';

export type CommandButton = Pick<Partial<ButtonProps>, 'size' | 'variant' | 'icon'> & {
  datasourceName?: string;
  deviceId: string;
  commandName: string;
  commandArgs: Array<{ name: string; value: any }>;
  buttonText: string;
  fullWidth: boolean;
  fullHeight: boolean;
  isButtonTextSetManually: boolean;
};

export type CommandButtonPanelProps = {
  commands: CommandButton[];
};
