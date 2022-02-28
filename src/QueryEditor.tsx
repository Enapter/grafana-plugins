import { defaults } from 'lodash';

import React, { PureComponent } from 'react';
import type * as monacoType from 'monaco-editor/esm/vs/editor/editor.api';
import { CodeEditor, Monaco } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onTextChange = (originalText: string) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, text: originalText });
    onRunQuery();
  };

  onEditorMount = (editor: monacoType.editor.IStandaloneCodeEditor, monaco: Monaco) => {
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      const text = editor.getValue();
      this.onTextChange(text);
    });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { text } = query;

    return (
      <>
        <CodeEditor
          height={150}
          showLineNumbers={true}
          onSave={this.onTextChange}
          onBlur={this.onTextChange}
          onEditorDidMount={this.onEditorMount}
          value={text}
          language={'yaml'}
          monacoOptions={{
            scrollBeyondLastLine: false,
            scrollBeyondLastColumn: 0,
            wordWrap: 'wordWrapColumn',
            wordWrapColumn: 100,
            wrappingIndent: 'same',
            minimap: { enabled: false },
            fontSize: 16,
          }}
        />
      </>
    );
  }
}
