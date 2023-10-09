import React, { useEffect } from 'react';
import { GrafanaTheme2, PanelOptionsEditorBuilder, SelectableValue } from '@grafana/data';
import { CommandButton, CommandButtonPanelProps } from '../types';
import {
  Button,
  CollapsableSection,
  Field,
  getAvailableIcons,
  InlineField,
  InlineFieldRow,
  Input,
  Label,
  Select,
  Switch,
  useStyles2,
} from '@grafana/ui';
import { useImmer } from 'use-immer';
import { css } from '@emotion/css';
import { getBackendSrv } from '@grafana/runtime';

const getStyles = (theme: GrafanaTheme2) => ({
  editorContainer: css`
    margin-top: ${theme.spacing(3)};
    > * + * {
      margin-top: ${theme.spacing(3)};
    }
  `,
  argumentsFieldsetContainer: css`
    > * + * {
      margin-top: ${theme.spacing(2)};
    }
  `,
});

const defaultValue: CommandButton[] = [
  {
    datasourceName: undefined,
    deviceId: '',
    commandName: '',
    commandArgs: [],
    variant: 'primary',
    icon: 'play',
    size: 'md',
    buttonText: '',
    fullHeight: false,
    fullWidth: true,
    isButtonTextSetManually: false,
  },
];

const buttonVariantOptions: Array<{ value: CommandButton['variant']; label: string }> = [
  { value: 'primary', label: 'Primary' },
  { value: 'secondary', label: 'Secondary' },
  { value: 'destructive', label: 'Destructive' },
  { value: 'success', label: 'Success' },
];

const buttonSizeOptions: Array<{ value: CommandButton['size']; label: string }> = [
  { value: 'sm', label: 'Small' },
  { value: 'md', label: 'Medium' },
  { value: 'lg', label: 'Large' },
];

const buttonIconOptions: Array<{ value: CommandButton['icon']; label: string }> = getAvailableIcons().map((icon) => ({
  value: icon as CommandButton['icon'],
  label: icon,
  icon,
}));

export const Editor: React.FC<{
  value: CommandButton[];
  onChange: (value: CommandButton[]) => void;
}> = ({ value, onChange }) => {
  const styles = useStyles2(getStyles);
  const [isButtonAppearanceOpen, setIsButtonAppearanceOpen] = React.useState(() => {
    return localStorage.getItem('enapter-commands-editor-button-appearance-open') === 'true';
  });
  const [datasources, setDatasources] = React.useState<Array<{ value: string; label: string }>>([]);
  const [commandDraft, setCommandDraft] = useImmer<CommandButton>(value[0]);
  const [invalidInputs, setInvalidInputs] = React.useState<Array<keyof CommandButton>>([]);

  const validateDeviceId = () => {
    if (commandDraft.deviceId) {
      setInvalidInputs((p) => p.filter((input) => input !== 'deviceId'));
      return true;
    }

    setInvalidInputs((p) => [...p, 'deviceId']);
    return false;
  };

  const validateCommandName = () => {
    if (commandDraft.commandName) {
      setInvalidInputs((p) => p.filter((input) => input !== 'commandName'));
      return true;
    }

    setInvalidInputs((p) => [...p, 'commandName']);
    return false;
  };

  const validateDatasourceName = () => {
    if (commandDraft.datasourceName) {
      setInvalidInputs((p) => p.filter((input) => input !== 'datasourceName'));
      return true;
    }

    setInvalidInputs((p) => [...p, 'datasourceName']);
    return false;
  };

  const validateInputs = () => {
    let isValid = true;
    isValid = validateDeviceId() && isValid;
    isValid = validateCommandName() && isValid;
    isValid = validateDatasourceName() && isValid;
    return isValid;
  };

  useEffect(() => {
    let canceled = false;

    async function getEnapterTelemetryDatasources() {
      let datasources: Array<{ value: string; label: string }> = [];
      try {
        const allDatasources = await getBackendSrv().get('/api/datasources');
        datasources = allDatasources
          .filter((ds: any) => ds.type.toLowerCase() === 'enapter-api')
          .map((ds: any) => ({ value: ds.name, label: ds.name }));
      } catch (e) {
        console.error(e);
      }
      if (canceled) {
        return;
      }
      setDatasources(datasources);
      setCommandDraft((draft) => {
        if (!draft.datasourceName && datasources.length) {
          draft.datasourceName = datasources[0].value;
        }
      });
    }

    getEnapterTelemetryDatasources();

    return () => {
      canceled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDatasourceNameChange = (option: SelectableValue<string>) => {
    setCommandDraft((draft) => {
      draft.datasourceName = option.value;
    });
  };

  const handleButtonNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCommandDraft((draft) => {
      draft.buttonText = e.target.value;
    });

    if (!commandDraft.isButtonTextSetManually) {
      setCommandDraft((draft) => {
        draft.isButtonTextSetManually = true;
      });
    }

    if (commandDraft.commandName === e.target.value) {
      setCommandDraft((draft) => {
        draft.isButtonTextSetManually = false;
      });
    }
  };

  const handleDeviceIdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCommandDraft((draft) => {
      draft.deviceId = e.target.value;
    });
  };

  const handleCommandNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCommandDraft((draft) => {
      draft.commandName = e.target.value;
    });

    if (!commandDraft.isButtonTextSetManually) {
      setCommandDraft((draft) => {
        draft.buttonText = String(e.target.value).trim();
      });
    }
  };

  const addCommandArg = () => {
    setCommandDraft((draft) => {
      draft.commandArgs.push({ name: '', value: '' });
    });
  };

  const removeCommandArg = (argIndex: number) => {
    setCommandDraft((draft) => {
      draft.commandArgs.splice(argIndex, 1);
    });
  };

  const handleArgNameChange = (e: React.ChangeEvent<HTMLInputElement>, argIndex: number) => {
    setCommandDraft((draft) => {
      draft.commandArgs[argIndex].name = e.target.value;
    });
  };

  const handleArgValueChange = (e: React.ChangeEvent<HTMLInputElement>, argIndex: number) => {
    setCommandDraft((draft) => {
      draft.commandArgs[argIndex].value = e.target.value;
    });
  };

  const handleButtonAppearanceToggle = () => {
    setIsButtonAppearanceOpen((p) => !p);
    localStorage.setItem('enapter-commands-editor-button-appearance-open', String(!isButtonAppearanceOpen));
  };

  const handleFullWidthChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCommandDraft((draft) => {
      draft.fullWidth = e.target.checked;
    });
  };

  const handleFullHeightChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCommandDraft((draft) => {
      draft.fullHeight = e.target.checked;
    });
  };

  const handleSizeChange = (option: SelectableValue<string>) => {
    setCommandDraft((draft) => {
      draft.size = option.value as CommandButton['size'];
    });
  };

  const handleVariantChange = (option: SelectableValue<string>) => {
    setCommandDraft((draft) => {
      draft.variant = option.value as CommandButton['variant'];
    });
  };

  const handleIconChange = (option: SelectableValue<string>) => {
    setCommandDraft((draft) => {
      draft.icon = option.value as CommandButton['icon'];
    });
  };

  useEffect(() => {
    if (
      !(
        commandDraft.fullWidth === value[0].fullWidth &&
        commandDraft.fullHeight === value[0].fullHeight &&
        commandDraft.size === value[0].size &&
        commandDraft.variant === value[0].variant &&
        commandDraft.icon === value[0].icon
      )
    ) {
      save();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    commandDraft.fullWidth,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    value[0].fullWidth,
    commandDraft.fullHeight,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    value[0].fullHeight,
    commandDraft.size,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    value[0].size,
    commandDraft.variant,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    value[0].variant,
    commandDraft.icon,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    value[0].icon,
  ]);

  const save = () => {
    onChange([commandDraft]);
  };

  const handleSave = () => {
    if (!validateInputs()) {
      return;
    }
    save();
  };

  return (
    <div className={styles.editorContainer}>
      <Field
        invalid={invalidInputs.includes('datasourceName')}
        error={'Required'}
        label="Enapter API data source name"
        required={true}
        disabled={datasources.length <= 1}
      >
        <Select
          onChange={handleDatasourceNameChange}
          onBlur={() => {
            if (validateDatasourceName()) {
              save();
            }
          }}
          id="datasourceName"
          value={
            datasources.filter((ds) => ds.value.toLowerCase() === commandDraft.datasourceName?.toLowerCase())[0] ||
            datasources[0] || { value: '', label: 'No Enapter API data sources found' }
          }
          options={datasources}
        />
      </Field>
      <Field invalid={invalidInputs.includes('deviceId')} error={'Required'} label="Device ID" required={true}>
        <Input
          id="deviceId"
          value={commandDraft.deviceId}
          onChange={handleDeviceIdChange}
          onBlur={() => {
            if (validateDeviceId()) {
              save();
            }
          }}
        />
      </Field>
      <Field
        invalid={invalidInputs.includes('commandName')}
        error={'Required'}
        label="Command name"
        description="Use command name from the device blueprint"
        required={true}
      >
        <Input
          id="commandName"
          value={commandDraft.commandName}
          onChange={handleCommandNameChange}
          onBlur={() => {
            if (validateCommandName()) {
              save();
            }
          }}
        />
      </Field>
      <Field label="Button text">
        <Input id="buttonText" value={commandDraft.buttonText} onChange={handleButtonNameChange} onBlur={save} />
      </Field>
      <div className={styles.argumentsFieldsetContainer}>
        <Label>Command arguments</Label>
        {commandDraft.commandArgs.map((arg, argIndex) => {
          return (
            <div key={`argument-fields-${argIndex}`}>
              <InlineFieldRow>
                <InlineField label="Name">
                  <Input
                    id={`argument-${argIndex}-name`}
                    value={arg.name}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleArgNameChange(e, argIndex)}
                  />
                </InlineField>
                <InlineField label="Value">
                  <Input
                    id={`argument-${argIndex}-value`}
                    value={arg.value}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleArgValueChange(e, argIndex)}
                  />
                </InlineField>
                <Button fill={'text'} variant={'destructive'} onClick={() => removeCommandArg(argIndex)}>
                  Remove
                </Button>
              </InlineFieldRow>
            </div>
          );
        })}
        <Button fill={'outline'} variant={'secondary'} style={{ marginTop: '1rem' }} onClick={addCommandArg}>
          Add argument
        </Button>
      </div>
      <CollapsableSection
        label={'Button appearance'}
        isOpen={isButtonAppearanceOpen}
        onToggle={handleButtonAppearanceToggle}
      >
        <Field label={'Full width'}>
          <Switch value={commandDraft.fullWidth} onChange={handleFullWidthChange} />
        </Field>
        <Field label={'Full height'}>
          <Switch value={commandDraft.fullHeight} onChange={handleFullHeightChange} />
        </Field>
        <Field label="Button size">
          <Select onChange={handleSizeChange} options={buttonSizeOptions} value={commandDraft.size} />
        </Field>
        <Field label="Button variant">
          <Select onChange={handleVariantChange} options={buttonVariantOptions} value={commandDraft.variant} />
        </Field>
        <Field label="Button icon">
          <Select onChange={handleIconChange} options={buttonIconOptions} value={commandDraft.icon} />
        </Field>
      </CollapsableSection>
      <Button onClick={handleSave}>Save</Button>
    </div>
  );
};

export function addEditor(builder: PanelOptionsEditorBuilder<CommandButtonPanelProps>) {
  builder.addCustomEditor({
    id: 'enapter-commands-editor',
    path: 'commands',
    name: 'Enapter Commands Editor',
    defaultValue,
    editor: (props) => <Editor value={props.value} onChange={props.onChange} />,
  });
}
