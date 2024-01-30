import React, { useEffect, useState } from 'react';
import { useDatasources } from '../hooks/useDatasources';
import { SelectableValue } from '@grafana/data';
import { Field, Select } from '@grafana/ui';
import { usePanel } from './PanelProvider';

export const DatasourceSelect: React.FC = () => {
  const {
    panel: { datasource },
    updatePanel,
  } = usePanel();

  const [options, setOptions] = useState<Array<SelectableValue<string>>>(() => {
    if (datasource) {
      return [{ label: datasource.name, value: datasource.uid }];
    }

    return [];
  });

  const [selected, setSelected] = useState(options[0]);

  useEffect(() => {
    if (selected?.value === datasource?.uid) {
      return;
    }

    if (selected?.value && selected?.label) {
      updatePanel((draft) => {
        draft.datasource = { uid: selected.value!, name: selected.label! };
      });
    }
  }, [selected, datasource, updatePanel]);

  const { isLoading } = useDatasources({
    onFound: (datasources) => {
      setOptions((existing) => {
        const dedupDatasources = datasources.filter(
          (ds) => !existing.find((e) => e.value === ds.uid)
        );

        const dedupDatasourcesOptions = dedupDatasources.map((ds) => ({
          label: ds.name,
          value: ds.uid,
        }));

        return [...existing, ...dedupDatasourcesOptions];
      });

      if (!selected) {
        setSelected({ label: datasources[0].name, value: datasources[0].uid });
      }
    },
  });

  return (
    <Field label="Datasource">
      <Select
        placeholder="Select Enapter API datasource"
        isLoading={isLoading}
        value={selected}
        options={options}
        onChange={setSelected}
      />
    </Field>
  );
};
