import { usePanel } from './PanelProvider';
import React from 'react';
import { AdvancedArgumentsFields } from './AdvancedArgumentsFields';
import { BasicArgumentsFields } from './BasicArgumentsFields';

export const CommandArguments = () => {
  const {
    panel: { currentCommand },
  } = usePanel();

  if (!currentCommand?.arguments) {
    return null;
  }

  return (
    <div key={currentCommand.key}>
      <div>
        {currentCommand.populateValuesCommand ? (
          <AdvancedArgumentsFields />
        ) : (
          <BasicArgumentsFields />
        )}
      </div>
    </div>
  );
};
