import { useCallback, useEffect, useRef } from 'react';
import { PanelState } from 'types/types';

const calculateFontSize = (width: number, textLength: number) => {
  return Math.max((width / (textLength + 2)) * 1.25, 16);
};

const shouldChangeSizeIfNarrow = (height: number, fontSize: number) => {
  return height < fontSize + Math.max(height / 7, 16);
};

const transformText = (button: HTMLButtonElement) => {
  if (!button) {
    return;
  }

  const innerTextLength = button.innerText.length;
  const buttonWidth = button.clientWidth;
  const buttonHeight = button.clientHeight;

  let fontSize = calculateFontSize(buttonWidth, innerTextLength);

  if (shouldChangeSizeIfNarrow(buttonHeight, fontSize)) {
    fontSize = buttonHeight / 1.5;
  }

  button.style.fontSize = `${fontSize}px`;

  button.style.setProperty('--ecp-button-font-size', `${fontSize}px`);

  const svgIcon = button.querySelector('svg');

  if (svgIcon) {
    if (svgIcon.parentElement) {
      svgIcon.parentElement.style.display = 'contents';
    }

    svgIcon.setAttribute('width', '1em');
    svgIcon.setAttribute('height', '1em');
    svgIcon.setAttribute('style', `margin-right: max(0.35em,8px)`);
  }

  const textContainer = button.querySelector('span');

  if (textContainer) {
    textContainer.style.display = 'contents';
    textContainer.setAttribute('style', `font-size: ${fontSize}px`);
  }
};

const resetText = (button: HTMLButtonElement) => {
  if (!button) {
    return;
  }

  button.removeAttribute('style');
  button.style.removeProperty('--ecp-button-font-size');

  const svgIcon = button.querySelector('svg');

  if (svgIcon) {
    if (svgIcon.parentElement) {
      svgIcon.parentElement.style.display = 'inline-block';
    }

    svgIcon.setAttribute('width', '16px');
    svgIcon.setAttribute('height', '16px');
    svgIcon.removeAttribute('style');
  }

  const textContainer = button.querySelector('span');

  if (textContainer) {
    textContainer.style.display = 'flex';
    textContainer.setAttribute('style', `font-size: ${textContainer.dataset.fontSize || '1rem'}`);
  }
};

export const useButtonTextResizer = (
  buttonRef: React.RefObject<HTMLButtonElement>,
  panelState: PanelState
) => {
  const buttonResizeObserverRef = useRef<ResizeObserver>(
    new ResizeObserver(() => {
      if (!buttonRef.current) {
        return;
      }

      transformText(buttonRef.current);
    })
  );

  const resetButtonText = useCallback(() => {
    if (!buttonRef.current) {
      return;
    }

    resetText(buttonRef.current);
  }, [buttonRef]);

  useEffect(() => {
    const ro = buttonResizeObserverRef.current;

    if (!buttonRef.current) {
      return;
    }

    if (panelState.appearance.shouldScaleText) {
      ro.observe(buttonRef.current);
    } else {
      ro.unobserve(buttonRef.current);
      resetButtonText();
    }

    return () => ro.disconnect();
  }, [buttonRef, panelState, resetButtonText]);
};
