import React from 'react';
import { DatasourceSelect } from './DatasourceSelect';
import { PanelProvider } from './PanelProvider';
import { CommandEditor } from './CommandEditor';
import { PanelState } from 'types/types';
import { StandardEditorProps } from '@grafana/data';
import { ApiClient } from 'api/client';
import { AppearanceEditor } from './AppearanceEditor';
import { migrateEditorV1ToV2 } from '../migrations/v1-to-v2';

export const apiClient = new ApiClient();

export const Editor: React.FC<StandardEditorProps<PanelState>> = ({ value, onChange }) => {
  return (
    <PanelProvider value={migrateEditorV1ToV2(value)} onChange={onChange}>
      <DatasourceSelect />
      <CommandEditor />
      <AppearanceEditor />
    </PanelProvider>
  );
};
