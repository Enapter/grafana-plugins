import { Field, Select, useStyles2 } from '@grafana/ui';
import { ConfirmationType } from '../types/types';
import React from 'react';
import { usePanel } from './PanelProvider';
import { GrafanaTheme2, SelectableValue } from '@grafana/data';
import { css } from '@emotion/css';

const getStyles = (theme: GrafanaTheme2) => {
  return {
    section: css({
      marginBlock: theme.spacing(4),
    }),
  };
};

export const AdvancedConfirmationTypeSelect = () => {
  const styles = useStyles2(getStyles);
  const { panel, updatePanel } = usePanel();

  const handleChange = (v: SelectableValue<ConfirmationType>) => {
    updatePanel((draft) => {
      if (!draft.currentCommand || v.value === undefined) {
        return;
      }

      draft.currentCommand.confirmationType = v.value;
    });
  };

  if (!panel.currentCommand) {
    return null;
  }

  return (
    <Field className={styles.section} label={'Open confirmation modal'}>
      <Select
        onChange={handleChange}
        options={[
          {
            label: 'Always',
            value: ConfirmationType.Always,
            description:
              'Always open the confirmation modal before running the command allowing to check or change arguments.',
          },
          {
            label: 'In case of invalid arguments',
            value: ConfirmationType.Invalid,
            description:
              'Run the command in the background, prepopulating arguments that have a "Populate" origin. Open the confirmation modal if there are invalid arguments present.',
          },
        ]}
        value={panel.currentCommand.confirmationType}
      />
    </Field>
  );
};
