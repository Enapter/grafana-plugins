import { getTemplateSrv } from '@grafana/runtime';

export const replaceVariables = (v: string) => getTemplateSrv().replace(v);
