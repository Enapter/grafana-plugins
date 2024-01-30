import React from 'react';
import { usePanel } from './PanelProvider';
import { BasicEditorArgumentField } from './BasicEditorArgumentField';

export const BasicArgumentsFields = () => {
  const {
    panel: { currentCommand },
  } = usePanel();

  if (!currentCommand?.arguments) {
    return null;
  }

  return (
    <>
      {Object.values(currentCommand.arguments).map((arg) => {
        return <BasicEditorArgumentField key={arg.key} arg={arg} />;
      })}
    </>
  );
};
