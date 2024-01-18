import React from 'react';
import { usePanel } from './PanelProvider';
import { Field, getAvailableIcons, Input, Select, Switch, useStyles2 } from '@grafana/ui';
import { css } from '@emotion/css';
import { PanelState } from '../types/types';
import { GrafanaTheme2, IconName } from '@grafana/data';
import { ColorPicker } from './ColorPicker';

const buttonIconOptions: Array<{ value: string; label: string }> = getAvailableIcons().map(
  (icon) => ({
    value: icon,
    label: icon,
    icon,
  })
);

const getStyles = (theme: GrafanaTheme2) => ({
  editorWrapper: css({
    border: 'none',
    borderTop: `1px solid ${theme.colors.border.weak}`,
    '> button:hover': {
      backgroundColor: theme.colors.background.secondary,
    },
    marginTop: theme.spacing(4),
  }),
  title: css({
    paddingTop: theme.spacing(4),
    paddingBottom: theme.spacing(2),
  }),
});

export const AppearanceEditor = () => {
  const styles = useStyles2(getStyles);

  const {
    panel: { appearance },
    updatePanel,
  } = usePanel();

  const handleChange = <T extends keyof PanelState['appearance']>(
    key: T,
    value: PanelState['appearance'][T]
  ) => {
    updatePanel((draft) => {
      draft.appearance[key] = value;
    });
  };

  return (
    <div className={styles.editorWrapper}>
      <h4 className={styles.title}>Appearance</h4>
      <Field label="Button text">
        <Input
          value={appearance.buttonText}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            handleChange('buttonText', e.target.value)
          }
        />
      </Field>
      <Field label="Background color">
        <ColorPicker
          color={appearance.bgColor}
          onChange={(color) => handleChange('bgColor', color)}
        />
      </Field>
      <Field label="Text color">
        <ColorPicker
          color={appearance.textColor}
          onChange={(color) => handleChange('textColor', color)}
        />
      </Field>
      <Field label={'Full width'}>
        <Switch
          value={appearance.fullWidth}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            handleChange('fullWidth', e.target.checked)
          }
        />
      </Field>
      <Field label={'Full height'}>
        <Switch
          value={appearance.fullHeight}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            handleChange('fullHeight', e.target.checked)
          }
        />
      </Field>
      <Field label={'Scale button text'}>
        <Switch
          value={appearance.shouldScaleText}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            handleChange('shouldScaleText', e.target.checked)
          }
        />
      </Field>
      <Field label="Button icon">
        <Select
          onChange={(v) => {
            handleChange('icon', (v.value || 'play') as IconName);
          }}
          options={buttonIconOptions}
          value={appearance.icon}
        />
      </Field>
    </div>
  );
};
