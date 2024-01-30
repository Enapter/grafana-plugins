import React from 'react';
import { Field } from '@grafana/ui';
import { Argument } from '../types/types';
import { ArgumentInput } from './ArgumentInput';
import { ArgValidator } from '../validation/arg-validator';
import { Draft } from 'immer';

type PanelArgumentFieldProps<T extends Argument = Argument> = {
  arg: T;
  onArgChange: (argKey: string, recipe: (draft: Draft<Argument>) => void) => void;
};

export const PanelArgumentField: React.FC<PanelArgumentFieldProps> = ({ arg, onArgChange }) => {
  const handleValueChange = (v: string | boolean) => {
    onArgChange(arg.key, (draft) => {
      draft.value = v;
    });
  };

  const handleBlur = () => {
    const argValidator = new ArgValidator(arg);

    onArgChange(arg.key, (draft) => {
      draft.isValid = argValidator.isValueValid();
      draft.errorMessage = argValidator.getErrorMessage();
    });
  };

  return (
    <Field
      key={arg.key}
      error={arg.errorMessage}
      invalid={!arg.isValid}
      label={arg.displayName}
      description={arg.description}
      required={arg.required}
      onBlur={handleBlur}
    >
      <ArgumentInput arg={arg} onChange={handleValueChange} />
    </Field>
  );
};
