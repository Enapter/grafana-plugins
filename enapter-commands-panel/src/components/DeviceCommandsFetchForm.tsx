import { Button, Field, Input, useStyles2 } from '@grafana/ui';
import { usePanel } from './PanelProvider';
import React, { useState } from 'react';
import { apiClient } from './Editor';
import { transformManifestCommandsToPanelCommands } from '../api/response-transformer';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { useNotificator } from '../hooks/useNotificator';
import { handleError } from '../utils/handleError';

const getStyles = (theme: GrafanaTheme2) => {
  return {
    section: css({
      marginBlock: theme.spacing(4),
    }),
  };
};

export const DeviceCommandsFetchForm = ({
  onLoadingStateChange,
}: {
  onLoadingStateChange: (state: boolean) => void;
}) => {
  const notificator = useNotificator();
  const styles = useStyles2(getStyles);
  const { panel, updatePanel } = usePanel();
  const [tempDeviceId, setTempDeviceId] = useState(panel.deviceId || '');

  const handleDeviceIdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTempDeviceId(e.target.value);
  };

  const handleGetCommands = async () => {
    if (!panel.datasource) {
      notificator.error('Datasource is not selected');

      return;
    }

    if (!tempDeviceId) {
      notificator.error('Device ID is not set');

      return;
    }

    const deviceId = String(tempDeviceId || '').trim();

    try {
      onLoadingStateChange(true);

      const { commands, errors } = await apiClient.fetchCommands(deviceId, panel.datasource.uid);

      if (errors.length) {
        notificator.error(errors.map((e: any) => e.message).join(', '));

        return;
      }

      const transformed = transformManifestCommandsToPanelCommands(commands);

      updatePanel((draft) => {
        draft.deviceId = deviceId;
        draft.commands = transformed;
        draft.currentCommand = undefined;

        if (commands) {
          draft.manifestCommands = commands;
        }
      });
    } catch (e) {
      handleError(e);

      updatePanel((draft) => {
        draft.deviceId = deviceId;
      });
    } finally {
      onLoadingStateChange(false);
    }
  };

  return (
    <Field
      className={styles.section}
      label="Device ID"
      description="Enter device ID to get available commands"
    >
      <div style={{ display: 'flex', gap: '0.5em' }}>
        <Input value={tempDeviceId} onChange={handleDeviceIdChange} />
        <Button variant={'secondary'} onClick={handleGetCommands}>
          Get commands
        </Button>
      </div>
    </Field>
  );
};
