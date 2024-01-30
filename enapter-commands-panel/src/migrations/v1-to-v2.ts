import { PanelState } from '../types/types';
import { V1CommandButton, V1CommandButtonPanelProps } from './types.v1';
import { defaultPanelState } from '../module';
import { PanelProps } from '@grafana/data';
import React from 'react';

export const isEditorV1 = (panel: any): panel is V1CommandButton[] => {
  try {
    return !!panel[0].deviceId;
  } catch (_) {
    return false;
  }
};

export const migrateEditorV1ToV2 = (panel: any): PanelState => {
  if (!panel) {
    return defaultPanelState;
  }

  if (isEditorV1(panel)) {
    return {
      ...defaultPanelState,
      deviceId: panel[0]?.deviceId,
      appearance: {
        ...defaultPanelState.appearance,
        icon: panel[0]?.icon || defaultPanelState.appearance.icon,
        buttonText: panel[0]?.buttonText || defaultPanelState.appearance.buttonText,
      },
    };
  }

  return panel;
};

const isPanelV1 = (
  panel: any
): panel is React.PropsWithChildren<PanelProps<V1CommandButtonPanelProps>> => {
  try {
    return !!panel?.options?.commands[0]?.deviceId;
  } catch (_) {
    return false;
  }
};

export const migratePanelV1ToV2 = (
  panel: any
): React.PropsWithChildren<PanelProps<{ commands: PanelState }>> => {
  if (!panel) {
    return panel;
  }

  if (isPanelV1(panel)) {
    return {
      ...panel,
      options: {
        commands: {
          ...defaultPanelState,
          deviceId: panel.options.commands[0].deviceId,
        },
      },
    } as unknown as React.PropsWithChildren<PanelProps<{ commands: PanelState }>>;
  }

  return panel;
};
