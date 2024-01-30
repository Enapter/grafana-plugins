import React, { useEffect, useState } from 'react';
import { DatasourceSelect } from './DatasourceSelect';
import { PanelProvider } from './PanelProvider';
import { CommandEditor } from './CommandEditor';
import { PanelState } from 'types/types';
import { StandardEditorProps } from '@grafana/data';
import { ApiClient } from 'api/client';
import { AppearanceEditor } from './AppearanceEditor';
import { isEditorV1, migrateEditorV1ToV2 } from '../migrations/v1-to-v2';
import { LoadingPlaceholder } from '@grafana/ui';

export const apiClient = new ApiClient();

export const Editor: React.FC<StandardEditorProps<PanelState>> = ({ value, onChange }) => {
  const [isMigrating, setIsMigrating] = useState(false);

  useEffect(() => {
    if (isEditorV1(value) && !isMigrating) {
      setIsMigrating(true);

      (async () => {
        onChange(migrateEditorV1ToV2(value));
        await new Promise((resolve) => setTimeout(resolve, 1500));
        setIsMigrating(false);
      })();
    }
  }, [isMigrating, onChange, value]);

  if (!value) {
    return null;
  }

  if (isMigrating || isEditorV1(value)) {
    return (
      <div style={{ padding: '2rem' }}>
        <LoadingPlaceholder text="Migrating to the new version..." />
      </div>
    );
  }

  return (
    <PanelProvider value={value} onChange={onChange}>
      <DatasourceSelect />
      <CommandEditor />
      <AppearanceEditor />
    </PanelProvider>
  );
};
