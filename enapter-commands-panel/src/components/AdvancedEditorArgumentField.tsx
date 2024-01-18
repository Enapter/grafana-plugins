import React from 'react';
import { InlineField, InlineFieldRow, Select, useStyles2 } from '@grafana/ui';
import { Argument, OriginType } from '../types/types';
import { ArgumentInput } from './ArgumentInput';
import { GrafanaTheme2, SelectableValue } from '@grafana/data';
import { css } from '@emotion/css';
import { usePanel } from './PanelProvider';
import { ArgValidator } from '../validation/arg-validator';

type AdvancedArgumentFieldProps<T extends Argument = Argument> = {
  arg: T;
};

const getStyles = (theme: GrafanaTheme2) => {
  return {
    container: css({
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(1),
      marginBottom: theme.spacing(3),
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
    fields: css({
      display: 'flex',
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing(0.5),
      '*': {
        minWidth: 0,
      },
    }),
    valueSource: css({
      flexBasis: '180px',
      '> div': {
        flexGrow: 1,
      },
      'div[class*="input-wrapper"]': {
        width: 'auto',
      },
    }),
    value: css({
      flexGrow: 1,
      '> div': {
        flexGrow: 1,
      },
      'div[class*="input-wrapper"]': {
        width: 'auto',
      },
    }),
  };
};

export const AdvancedEditorArgumentField: React.FC<AdvancedArgumentFieldProps> = ({ arg }) => {
  const { updatePanel } = usePanel();
  const styles = useStyles2(getStyles);

  const handleValueChange = (v: string | boolean) => {
    updatePanel((draft) => {
      if (!draft.currentCommand || !draft.currentCommand.arguments?.[arg.key]) {
        return;
      }

      draft.currentCommand.arguments[arg.key].value = v;
    });
  };

  const handleBlur = () => {
    const argValidator = new ArgValidator(arg, { skip: { required: true } });

    updatePanel((draft) => {
      if (!draft.currentCommand || !draft.currentCommand.arguments?.[arg.key]) {
        return;
      }

      const argRef = draft.currentCommand.arguments[arg.key];
      argRef.isValid = argValidator.isValueValid();
      argRef.errorMessage = argValidator.getErrorMessage();
    });
  };

  const handleOriginTypeChange = (v: SelectableValue<OriginType>) => {
    updatePanel((draft) => {
      if (!draft.currentCommand || !draft.currentCommand.arguments?.[arg.key]) {
        return;
      }

      draft.currentCommand.arguments[arg.key].originType = v.value ?? OriginType.Populate;
    });
  };

  return (
    <div className={styles.container}>
      <div className={styles.label} style={{ gridColumn: '1 / 3' }}>
        <div>
          {arg.displayName}
          {arg.required && <span>&nbsp;*</span>}
        </div>
        <span className={styles.description}>{arg.description}</span>
      </div>
      <InlineFieldRow className={styles.fields}>
        <InlineField label={'Origin'} className={styles.valueSource}>
          <Select
            onChange={handleOriginTypeChange}
            options={[
              {
                label: 'Populate',
                value: OriginType.Populate,
                description:
                  'Populate the field with a value from the command specified in the “populate_values_command” field in the blueprint. If this populate command returns no value, use the value set in the “Value” field as a fallback.',
              },
              {
                label: 'Fixed value',
                value: OriginType.Fixed,
                description: 'Do not populate the value. Use the value set in the “Value” field.',
              },
            ]}
            value={arg.originType ?? OriginType.Populate}
            width={10}
          />
        </InlineField>
        <InlineField label={'Value'} className={styles.value} onBlur={handleBlur}>
          <ArgumentInput arg={arg} onChange={handleValueChange} />
        </InlineField>
      </InlineFieldRow>
    </div>
  );
};
