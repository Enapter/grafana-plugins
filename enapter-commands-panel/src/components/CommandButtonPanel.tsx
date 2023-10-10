import React, { useCallback, useState } from 'react';
import { AppEvents, PanelProps } from '@grafana/data';
import { Button, ConfirmModal } from '@grafana/ui';
import { CommandButtonPanelProps } from '../types';
import { getAppEvents, getBackendSrv, getTemplateSrv } from '@grafana/runtime';

// if (process.env.NODE_ENV === 'development') {
//   console.log('Setting up mirage server');
//   require('../mocks/mirage-server').setupMirage();
// }

const getArgumentsAsObject = (
  args: Array<{
    name: string;
    value: any;
  }>
) => {
  const result: { [key: string]: any } = {};
  args.forEach((arg) => {
    let value: string | number = getTemplateSrv().replace(arg.value)

    try {
      const isContainCommaOrDot = value.includes(',') || value.includes('.');
      const isInt = Number.isInteger(Number.parseInt(value, 10));
      const isFloat = Number.isFinite(Number.parseFloat(value));

      if (!isContainCommaOrDot && isInt) {
        value = Number.parseInt(value, 10);
      } else if (isFloat) {
        value = Number.parseFloat(value);
      }
    } catch (e) {
      console.error(e);
    }

    result[arg.name] = value;
  });
  return result;
};

const replaceVariables = (v: string) => getTemplateSrv().replace(v)

export const CommandButtonPanel: React.FC<PanelProps<CommandButtonPanelProps>> = (props) => {
  const { commandName, commandArgs, buttonText, size, variant, icon, fullWidth, fullHeight, deviceId, datasourceName } = (() => {
    const command = props.options.commands[0]

    return {
      ...command,
      commandName: replaceVariables(command.commandName),
      commandArgs: getArgumentsAsObject(command.commandArgs),
      buttonText: replaceVariables(command.buttonText),
      deviceId: replaceVariables(command.deviceId)
    }
  })()

  const [commandState, setCommandState] = useState<'idle' | 'running'>('idle');

  const [isConfirmationOpen, setIsConfirmationOpen] = useState(false);

  const areBaseCommandArgsPresent = useCallback(() => {
    const appEvents = getAppEvents();

    if (!deviceId) {
      appEvents.publish({ type: AppEvents.alertError.name, payload: [`Device ID is required`] });
      return false;
    }

    if (!commandName) {
      appEvents.publish({ type: AppEvents.alertError.name, payload: [`Command name is required`] });
      return false;
    }

    if (!datasourceName) {
      appEvents.publish({ type: AppEvents.alertError.name, payload: [`Data source name is required`] });
      return false;
    }

    return true;
  }, [commandName, datasourceName, deviceId]);

  const handleSendCommand = () => {
    if (!areBaseCommandArgsPresent()) {
      return;
    }

    setIsConfirmationOpen(true);
  };

  const sendCommand = useCallback(async () => {
    setIsConfirmationOpen(false);

    const appEvents = getAppEvents();
    let uid;
    try {
      uid = await getBackendSrv()
        .get(`/api/datasources/name/${datasourceName}`)
        .then((res) => res.uid);
    } catch (e) {
      console.error(e);
      appEvents.publish({ type: AppEvents.alertError.name, payload: [`Data source ${datasourceName} not found`] });
      return;
    }

    if (!uid) {
      appEvents.publish({ type: AppEvents.alertError.name, payload: [`Data source ${datasourceName} not found`] });
      return;
    }

    const url = '/api/ds/query';
    const body = {
      queries: [
        {
          queryType: 'command',
          refId: 'A',
          datasource: {
            uid,
          },
          payload: {
            commandName,
            commandArgs,
            deviceId,
          },
        },
      ],
    };

    try {
      setCommandState('running');
      const res = await getBackendSrv().post(url, body);
      const grafanaBackendResponse: any = res.results.A;
      const grafanaApiError = grafanaBackendResponse.error;

      if (grafanaApiError) {
        appEvents.publish({
          type: AppEvents.alertError.name,
          payload: [`Command ${commandName} failed`, grafanaApiError],
        });
      } else {
        const telemetryBackendResponse = grafanaBackendResponse.frames[0].data;
        const state = telemetryBackendResponse.values[0][0];
        const errors = telemetryBackendResponse.values[1];

        switch (state) {
          case 'succeeded':
            appEvents.publish({
              type: AppEvents.alertSuccess.name,
              payload: [`Command ${commandName} succeeded`],
            });
            break;
          case 'error':
          case 'platform_error':
            appEvents.publish({
              type: AppEvents.alertError.name,
              payload: [`Command ${commandName} failed`, errors.map((e: any) => e.message).join('. ')],
            });
            break;
          default:
            break;
        }
      }
    } catch (e) {
      console.error(e);
      appEvents.publish({
        type: AppEvents.alertError.name,
        payload: ['Something went wrong'],
      });
    } finally {
      setCommandState('idle');
    }
  }, [deviceId, commandName, datasourceName, commandArgs]);

  return (
    <>
      <Button
        size={size}
        variant={variant}
        icon={commandState === 'running' ? 'fa fa-spinner' : icon}
        onClick={handleSendCommand}
        disabled={commandState === 'running'}
        style={{
          width: fullWidth ? '100%' : undefined,
          height: fullHeight ? '100%' : undefined,
          whiteSpace: 'pre-wrap',
        }}
      >
        {buttonText}
      </Button>
      <ConfirmModal
        isOpen={isConfirmationOpen}
        title={`Run ${commandName}`}
        body={`Are you sure you want to run ${commandName}?`}
        confirmText="Run"
        icon="exclamation-triangle"
        onConfirm={sendCommand}
        onDismiss={() => setIsConfirmationOpen(false)}
      />
    </>
  );
};
