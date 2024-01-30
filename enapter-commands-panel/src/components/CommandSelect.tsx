import React, { useMemo } from 'react';
import { Field, Select, useStyles2 } from '@grafana/ui';
import { usePanel } from './PanelProvider';
import { GrafanaTheme2, SelectableValue } from '@grafana/data';
import { css } from '@emotion/css';

const getStyles = (theme: GrafanaTheme2) => {
  return {
    commandSelect: css({
      marginBlock: theme.spacing(4),
    }),
    populatingLoader: css({
      marginBlock: theme.spacing(4),
    }),
    label: css({
      fontSize: theme.typography.bodySmall.fontSize,
      fontWeight: theme.typography.fontWeightMedium,
      lineHeight: 1.25,
      marginBottom: theme.spacing(0.5),
      color: theme.colors.text.primary,
      maxWidth: '480px',
    }),
    description: css({
      color: theme.colors.text.secondary,
      fontSize: theme.typography.bodySmall.fontSize,
      fontWeight: theme.typography.fontWeightRegular,
    }),
  };
};

export const CommandSelect: React.FC = () => {
  const styles = useStyles2(getStyles);

  const {
    panel: { commands, currentCommand },
    updatePanel,
  } = usePanel();

  const options = useMemo(() => {
    return Object.entries(commands)
      .sort((a, b) => {
        const first = a[1].displayName.toLowerCase();
        const second = b[1].displayName.toLowerCase();

        return first.localeCompare(second);
      })
      .map(([commandName, command]) => ({
        label: command.displayName,
        value: commandName,
      }));
  }, [commands]);

  const handleOnChange = async (v: SelectableValue<string>) => {
    updatePanel((draft) => {
      draft.currentCommand = draft.commands[v.value!];
    });
  };

  return (
    <div className={styles.commandSelect}>
      <Field label="Command" description="Select command to execute">
        <Select options={options} value={currentCommand?.key} onChange={handleOnChange} />
      </Field>
      {currentCommand?.description && (
        <div>
          <div className={styles.label}>Command description</div>
          <span className={styles.description}>{currentCommand.description}</span>
        </div>
      )}
    </div>
  );
};
