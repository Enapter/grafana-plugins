import { Checkbox, Field, useStyles2 } from '@grafana/ui';
import { usePanel } from './PanelProvider';
import React from 'react';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { ConfirmationType } from '../types/types';

const getStyles = (theme: GrafanaTheme2) => {
  return {
    section: css({
      marginBlock: theme.spacing(4),
    }),
  };
};

export const BasicConfirmationTypeCheckbox = () => {
  const styles = useStyles2(getStyles);
  const { panel, updatePanel } = usePanel();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    updatePanel((draft) => {
      if (!draft.currentCommand) {
        return;
      }

      draft.currentCommand.confirmationType = e.target.checked
        ? ConfirmationType.Always
        : ConfirmationType.Never;
    });
  };

  if (!panel.currentCommand) {
    return null;
  }

  const isChecked = panel.currentCommand.confirmationType === ConfirmationType.Always;

  return (
    <Field
      className={styles.section}
      label={'Run with confirmation'}
      disabled={!!panel.currentCommand.confirmation}
      description={'Can not be unchecked for commands that require confirmation in the blueprint.'}
    >
      <Checkbox checked={isChecked} onChange={handleChange} />
    </Field>
  );
};
