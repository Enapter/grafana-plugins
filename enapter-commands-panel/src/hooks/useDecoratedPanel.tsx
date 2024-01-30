import { GrafanaTheme2, PanelProps } from '@grafana/data';
import { PanelState } from '../types/types';
import { capitalize } from '../utils/capitalize';
import { css } from '@emotion/css';
import { lightenColor } from '../utils/lightenColor';
import { useStyles2 } from '@grafana/ui';
import { useButtonTextResizer } from './useButtonTextResizer';
import { replaceVariables } from '../utils/replaceVariables';

type StylesProps = Pick<
  PanelState['appearance'],
  'bgColor' | 'textColor' | 'fullWidth' | 'fullHeight' | 'shouldScaleText'
>;

const getStyles = (
  _: GrafanaTheme2,
  { bgColor, textColor, fullWidth, fullHeight, shouldScaleText }: StylesProps
) => {
  return {
    panel: css({
      width: '500px',
    }),
    commandButton: css({
      backgroundColor: bgColor,
      color: textColor,
      width: fullWidth ? '100%' : 'auto',
      height: fullHeight ? '100%' : 'auto',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      '&:hover': {
        backgroundColor: lightenColor(bgColor, 12),
        color: textColor,
      },
      '&:active': {
        backgroundColor: lightenColor(bgColor, 20),
        color: textColor,
      },
      '&:disabled': {
        backgroundColor: bgColor,
        color: textColor,
        filter: 'grayscale(35%)',
        opacity: 0.75,
      },
      span: {
        lineHeight: shouldScaleText ? 'max(150%, 30px)' : '30px',
        whiteSpace: 'break-spaces',
      },
      svg: {
        minWidth: '1em',
        minHeight: '1em',
      },
      'i.fa.fa-spinner': {
        maxWidth: '1em',
        maxHeight: '1em',
        marginRight: 'max(0.35em,8px)',
      },
    }),
  };
};

export const useDecoratedPanel = (
  props: PanelProps<{ commands: PanelState }>,
  buttonRef: React.RefObject<HTMLButtonElement>
) => {
  const config = props.options.commands;
  useButtonTextResizer(buttonRef, config);

  const styles = useStyles2((theme) =>
    getStyles(theme, {
      bgColor: config.appearance.bgColor,
      textColor: config.appearance.textColor,
      fullWidth: config.appearance.fullWidth,
      fullHeight: config.appearance.fullHeight,
      shouldScaleText: config.appearance.shouldScaleText,
    })
  );

  const modalTitle = capitalize(config.currentCommand?.displayName || 'run command');
  const icon = config.appearance.icon;

  const buttonText = capitalize(
    replaceVariables(
      config.appearance.buttonText || config.currentCommand?.displayName || 'run command'
    )
  );

  return {
    modalTitle,
    panelButtonClassName: styles.commandButton,
    panelClassName: styles.panel,
    icon,
    buttonText,
  };
};
