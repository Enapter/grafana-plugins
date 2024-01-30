import React from 'react';
import { Checkbox, Input, Select, useStyles2 } from '@grafana/ui';
import { Argument } from '../types/types';
import _toString from 'lodash/toString';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';

type ArgumentInputProps = {
  arg: Argument;
  onChange: (value: string | boolean) => void;
  width?: number;
  disabled?: boolean;
  isAlwaysAsk?: boolean;
};

const getStyles = (theme: GrafanaTheme2) => {
  return {
    container: css({
      flexGrow: 1,
    }),
  };
};

export const ArgumentInput: React.FC<ArgumentInputProps> = ({
  arg,
  onChange,
  width,
  disabled,
  isAlwaysAsk,
}) => {
  const styles = useStyles2(getStyles);

  if ('options' in arg && arg.options.length) {
    return (
      <Select
        className={styles.container}
        options={arg.options}
        value={arg.value ?? arg.options.find((o) => o.value === arg.defaultValue)}
        onChange={(o) => {
          if (o === null) {
            onChange('');

            return;
          }

          onChange(_toString(o.value));
        }}
        width={width}
        isClearable={true}
      />
    );
  }

  if (arg.type === 'boolean') {
    return (
      <Checkbox
        className={styles.container}
        value={arg.value ?? arg.defaultValue}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange(e.target.checked)}
        width={width}
      />
    );
  }

  return (
    <div>
      <Input
        className={styles.container}
        type="text"
        value={arg.value ?? arg.defaultValue}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        width={width}
        disabled={disabled}
        placeholder={
          disabled ? 'Auto-populated' : isAlwaysAsk ? 'Enter default value' : 'Enter value'
        }
      />
    </div>
  );
};
