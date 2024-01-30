import { useEffect, useState } from 'react';
import { PanelState } from 'types/types';
import { apiClient } from '../components/Editor';
import { handleError } from 'utils/handleError';

type Datasource = NonNullable<PanelState['datasource']>;
type UseDatasourcesHookProps = {
  onFound: (datasources: Datasource[]) => void;
};

export const useDatasources = ({ onFound }: UseDatasourcesHookProps) => {
  const [isLoading, setIsLoading] = useState(true);
  const [datasources, setDatasources] = useState<Datasource[]>([]);

  useEffect(() => {
    let canceled = false;

    async function getEnapterTelemetryDatasources() {
      setIsLoading(true);
      let datasources: Datasource[] = [];

      try {
        const allDatasources = await apiClient.fetchDatasources();

        datasources = allDatasources
          .filter((ds) => ds.type.toLowerCase() === 'enapter-api')
          .map<Datasource>((ds) => ({ uid: ds.uid, name: ds.name }));

        if (datasources.length) {
          onFound(datasources);
        }
      } catch (e) {
        handleError(e);
      } finally {
        setIsLoading(false);
      }

      if (canceled) {
        return;
      }

      setDatasources(datasources);
    }

    getEnapterTelemetryDatasources();

    return () => {
      canceled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { datasources, isLoading };
};
