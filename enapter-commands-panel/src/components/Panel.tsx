import React, { useEffect, useRef } from 'react';
import { PanelProps } from '@grafana/data';
import { Alert, Button, LoadingPlaceholder, Modal } from '@grafana/ui';
import { ConfirmationType, OriginType, PanelState } from '../types/types';
import { useNotificator } from '../hooks/useNotificator';
import { useModalControls } from '../hooks/useModalControls';
import { useDecoratedPanel } from '../hooks/useDecoratedPanel';
import { useNotPersistedArgs } from '../hooks/useNotPersistedArgs';
import { MisconfiguredPanelError, useRunCommandWorkflow } from '../hooks/useRunCommandWorkflow';
import { stringify } from '../utils/stringify';
import { handleError } from '../utils/handleError';
import { PanelArgumentField } from './PanelArgumentField';
import { migratePanelV1ToV2 } from '../migrations/v1-to-v2';

export const Panel: React.FC<PanelProps<{ commandButton: PanelState }>> = (props) => {
  const data = migratePanelV1ToV2(props);
  const notificator = useNotificator();

  const { currentCommand, datasource, deviceId } = data.options.commandButton;

  const { isOpen, openModal, closeModal } = useModalControls();

  const buttonRef = useRef<HTMLButtonElement>(null);

  const { modalTitle, icon, panelButtonClassName, panelClassName, buttonText } = useDecoratedPanel(
    data,
    buttonRef
  );

  const { args, updateArg, hasInvalidArgs, resetArgs, validateArgs } = useNotPersistedArgs(data);

  const { runCommand, isRunning, populateValues, isPopulating, populateValuesAndRunCommand } =
    useRunCommandWorkflow({
      deviceId,
      datasource,
    });

  const handlePanelButtonClick = async () => {
    if (!currentCommand) {
      notificator.error('Failed to run the command', 'Command is not set');

      return;
    }

    if (
      currentCommand.confirmationType === ConfirmationType.Never ||
      currentCommand.confirmationType === ConfirmationType.Invalid
    ) {
      try {
        const payload = await populateValuesAndRunCommand(currentCommand, args);

        notificator.success(
          `Command "${currentCommand.displayName}" succeeded`,
          stringify(payload)
        );

        return;
      } catch (e: unknown) {
        handleError(e);

        if (!(e instanceof MisconfiguredPanelError)) {
          return;
        }
      }
    }

    openModal();

    try {
      if (currentCommand.populateValuesCommand) {
        const payload = await populateValues(currentCommand.populateValuesCommand);

        Object.entries(payload).forEach(([argName, argValue]) => {
          updateArg(argName, (draft) => {
            if (draft.originType === OriginType.Populate) {
              draft.value = stringify(argValue);
            }
          });
        });

        validateArgs();
      }
    } catch (e: unknown) {
      handleError(e, 'Error occurred during values prepopulation');
    }
  };

  const handleRunCommand = async () => {
    if (!currentCommand) {
      return;
    }

    if (hasInvalidArgs) {
      notificator.error(
        `Failed to run the "${currentCommand.displayName}" command`,
        'Some arguments are invalid. Please check them and try again.'
      );

      return;
    }

    closeModal();

    try {
      const payload = await runCommand(currentCommand.key, args);
      notificator.success(`Command "${currentCommand.displayName}" succeeded`, stringify(payload));
    } catch (e: unknown) {
      handleError(e);
    }
  };

  useEffect(() => {
    if (!isOpen) {
      resetArgs();
    }
  }, [isOpen, resetArgs]);

  return (
    <>
      <Button
        ref={buttonRef}
        className={panelButtonClassName}
        icon={isRunning ? 'fa fa-spinner' : icon}
        disabled={isRunning}
        onClick={handlePanelButtonClick}
      >
        {buttonText}
      </Button>
      <Modal
        className={panelClassName}
        isOpen={isOpen}
        onClickBackdrop={closeModal}
        onDismiss={closeModal}
        title={modalTitle}
      >
        {isPopulating ? (
          <LoadingPlaceholder text={'Populating arguments...'} />
        ) : (
          <>
            <p>{currentCommand?.description}</p>
            {currentCommand?.confirmation && (
              <>
                <Alert
                  severity={currentCommand.confirmation.severity}
                  title={currentCommand.confirmation.title}
                >
                  {currentCommand.confirmation.description}
                </Alert>
              </>
            )}
            {currentCommand?.arguments && (
              <>
                <p>You are running the command with arguments:</p>
                {Object.values(args).map((arg) => {
                  return <PanelArgumentField key={arg.key} arg={arg} onArgChange={updateArg} />;
                })}
              </>
            )}
            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <Button onClick={handleRunCommand} disabled={isRunning}>
                Run command
              </Button>
            </div>
          </>
        )}
      </Modal>
    </>
  );
};
