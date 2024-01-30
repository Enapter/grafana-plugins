import React from 'react';
import { AdvancedEditorArgumentField } from './AdvancedEditorArgumentField';
import { Button, Dropdown, Icon, useStyles2 } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { usePanel } from './PanelProvider';

const getStyles = (theme: GrafanaTheme2) => {
  return {
    alert: css({
      display: 'flex',
      alignItems: 'flex-start',
      gap: theme.spacing(1),
      marginBottom: theme.spacing(4),
      backgroundColor: theme.colors.background.secondary,
      borderRadius: theme.shape.borderRadius(1),
      padding: theme.spacing(2),
      maxWidth: '480px',
    }),
    alertDescription: css({
      color: theme.colors.text.secondary,
      fontSize: theme.typography.bodySmall.fontSize,
      fontWeight: theme.typography.fontWeightRegular,
      marginBottom: theme.spacing(1),
    }),
    learnMoreContainer: css({
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(1),
      padding: theme.spacing(2),
      maxWidth: '400px',
      height: 'max-content',
      fontSize: theme.typography.bodySmall.fontSize,
      fontWeight: theme.typography.fontWeightRegular,
      color: theme.colors.text.secondary,
      backgroundColor: theme.colors.background.secondary,
      borderRadius: theme.shape.borderRadius(2),
      boxShadow: theme.shadows.z3,
      ul: {
        paddingLeft: theme.spacing(3),
      },
    }),
    learnMoreTitle: css({
      fontSize: theme.typography.bodySmall.fontSize,
      fontWeight: theme.typography.fontWeightMedium,
      lineHeight: 1.25,
      marginBottom: theme.spacing(0.5),
      color: theme.colors.text.primary,
    }),
  };
};

export const AdvancedArgumentsFields = () => {
  const {
    panel: { currentCommand },
  } = usePanel();

  if (!currentCommand?.arguments) {
    return null;
  }

  return (
    <>
      <AdvancedEditorAlert />
      {Object.values(currentCommand.arguments).map((arg) => {
        return <AdvancedEditorArgumentField arg={arg} key={arg.key} />;
      })}
    </>
  );
};

const LearnMoreContent = () => {
  const styles = useStyles2(getStyles);

  return (
    <div className={styles.learnMoreContainer}>
      <div className={styles.learnMoreTitle}>Command&apos;s arguments editor variants</div>
      <div>
        Two variants of the command&apos;s arguments editor are available:
        <ul>
          <li>Basic</li>
          <li>Advanced</li>
        </ul>
      </div>
      <div>
        Basic editor is used when the &ldquo;populate_values_command&rdquo; field is not set in the
        blueprint.
      </div>
      <div>
        Advanced editor is used when the &ldquo;populate_values_command&rdquo; field is set in the
        blueprint. You may control how the arguments are populated by the populate command. Use the
        &ldquo;Origin&rdquo; field to select the source of the argument&apos;s value.
      </div>
      <div>
        When the &ldquo;Origin&rdquo; field is set to &ldquo;Populate&rdquo;, the argument&apos;s
        value is filled in by the command specified in the &ldquo;populate_values_command&rdquo;
        field within the blueprint. If this populate command returns no value, the value set in the
        &ldquo;Value&rdquo; field is used as a fallback.
      </div>
      <div>
        Conversely, when the &ldquo;Origin&rdquo; field is set to &ldquo;Fixed value&rdquo;, the
        argument&apos;s value is not populated. Instead, the value set in the &ldquo;Value&rdquo;
        field is used.
      </div>
      <div>
        Read further at{' '}
        <a
          rel="noreferrer"
          target="_blank"
          href="https://developers.enapter.com/docs/reference/#command-prepopulation"
        >
          developers.enapter.com
        </a>
        .
      </div>
    </div>
  );
};

const AdvancedEditorAlert = () => {
  const styles = useStyles2(getStyles);

  return (
    <div className={styles.alert}>
      <div>
        <Icon name={'info-circle'} />
      </div>
      <div>
        <div>Advanced arguments editor</div>
        <div className={styles.alertDescription}>
          You cannot use the basic arguments editor when the &ldquo;populate_values_command&rdquo;
          field is set for this command in the blueprint.
        </div>
        <Dropdown overlay={LearnMoreContent}>
          <Button variant={'secondary'} size={'sm'} fill={'outline'}>
            Learn more
          </Button>
        </Dropdown>
      </div>
    </div>
  );
};
