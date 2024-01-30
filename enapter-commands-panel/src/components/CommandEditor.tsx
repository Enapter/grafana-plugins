import React, { useState } from 'react';
import { usePanel } from './PanelProvider';
import { LoadingPlaceholder } from '@grafana/ui';
import { CommandArguments } from './CommandArguments';
import { CommandSelect } from './CommandSelect';
import { BasicConfirmationTypeCheckbox } from './BasicConfirmationTypeCheckbox';
import { DeviceCommandsFetchForm } from './DeviceCommandsFetchForm';
import { AdvancedConfirmationTypeSelect } from './AdvancedConfirmationTypeSelect';
import { PanelState } from '../types/types';

export const CommandEditor = () => {
  const [isLoading, setIsLoading] = useState(false);
  const { panel } = usePanel();

  return (
    <div>
      <DeviceCommandsFetchForm onLoadingStateChange={setIsLoading} />
      {isLoading && <LoadingPlaceholder text={'Loading commands...'} />}
      {panel.deviceId && !isLoading && <CommandSelect />}
      {panel.currentCommand && !isLoading && <CommandArguments />}
      {panel.currentCommand && !isLoading && (
        <ConfirmationTypeSelect command={panel.currentCommand} />
      )}
    </div>
  );
};

const ConfirmationTypeSelect = ({
  command,
}: {
  command: NonNullable<PanelState['currentCommand']>;
}) => {
  if (command.populateValuesCommand) {
    return <AdvancedConfirmationTypeSelect />;
  }

  return <BasicConfirmationTypeCheckbox />;
};
