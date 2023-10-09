import { PanelPlugin } from '@grafana/data';
import { CommandButtonPanel } from './components/CommandButtonPanel';
import { addEditor } from './components/Editor';
import { CommandButtonPanelProps } from './types';

export const plugin = new PanelPlugin<CommandButtonPanelProps>(CommandButtonPanel);

plugin.setPanelOptions((builder) => addEditor(builder));
