import { StandardEditorProps } from '@grafana/data';
import React, { createContext, useContext } from 'react';
import { Argument, PanelState } from 'types/types';
import { Draft, produce } from 'immer';

type PanelContextType = Pick<StandardEditorProps<PanelState>, 'value' | 'onChange'>;

type PanelProviderType = {
  panel: PanelState;
  updatePanel: (recipe: (draft: Draft<PanelState>) => void) => void;
  updateArg: (argKey: string, recipe: (draft: Draft<Argument>) => void) => void;
};

const PanelContext = createContext<PanelProviderType>({} as PanelProviderType);

export const PanelProvider: React.FC<PanelContextType> = ({ children, value, onChange }) => {
  const valueRef = React.useRef(value);
  valueRef.current = value;

  const updatePanel = (recipe: (draft: Draft<PanelState>) => void) => {
    onChange(produce(valueRef.current, recipe));
  };

  const updateArg = (argKey: string, recipe: (draft: Draft<Argument>) => void) => {
    updatePanel((draft) => {
      if (!draft.currentCommand || !draft.currentCommand.arguments) {
        return;
      }

      draft.currentCommand.arguments[argKey] = produce(
        draft.currentCommand.arguments[argKey],
        recipe
      );
    });
  };

  return (
    <PanelContext.Provider value={{ panel: value, updatePanel, updateArg }}>
      {children}
    </PanelContext.Provider>
  );
};

export const usePanel = () => {
  return useContext(PanelContext);
};
