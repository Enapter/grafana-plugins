import { PanelPlugin } from '@grafana/data';
import { Panel } from './components/Panel';
import { Editor } from './components/Editor';
import { PanelState } from 'types/types';

export const defaultPanelState: PanelState = {
  pluginVersion: '1.0',
  deviceId: '',
  commands: {},
  manifestCommands: {},
  appearance: {
    icon: 'play',
    bgColor: '#3871dc',
    textColor: '#ffffff',
    fullWidth: false,
    fullHeight: false,
    shouldScaleText: false,
  },
};

export const plugin = new PanelPlugin(Panel);

plugin.setPanelOptions((builder) =>
  builder.addCustomEditor({
    id: 'enapter-commands-editor',
    path: 'commandButton',
    name: '',
    defaultValue: defaultPanelState,
    editor: Editor,
  })
);
