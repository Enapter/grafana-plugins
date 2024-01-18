import { PanelState } from '../types/types';
import { V1CommandButtonPanelProps } from './types.v1';
import { defaultPanelState } from '../module';
import { PanelProps } from '@grafana/data';
import React from 'react';

const isEditorV1 = (panel: any): panel is V1CommandButtonPanelProps => {
  try {
    return !!panel?.commands[0]?.deviceId;
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
      deviceId: panel.commands[0].deviceId,
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
): React.PropsWithChildren<PanelProps<{ commandButton: PanelState }>> => {
  if (!panel) {
    return {
      ...defaultPanelState,
      options: {
        commandButton: defaultPanelState,
      },
    } as unknown as React.PropsWithChildren<PanelProps<{ commandButton: PanelState }>>;
  }

  if (isPanelV1(panel)) {
    return {
      ...panel,
      options: {
        commandButton: {
          ...defaultPanelState,
          deviceId: panel.options.commands[0].deviceId,
        },
      },
    } as unknown as React.PropsWithChildren<PanelProps<{ commandButton: PanelState }>>;
  }

  return panel;
};
