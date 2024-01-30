import React from 'react';
import { Field } from '@grafana/ui';
import { Argument } from '../types/types';
import { ArgumentInput } from './ArgumentInput';
import { ArgValidator } from '../validation/arg-validator';
import { usePanel } from './PanelProvider';

type BasicArgumentFieldProps<T extends Argument = Argument> = {
  arg: T;
};

export const BasicEditorArgumentField: React.FC<BasicArgumentFieldProps> = ({ arg }) => {
  const { updatePanel } = usePanel();

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
