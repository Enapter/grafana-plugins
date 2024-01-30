import { ColorPicker as GrafanaColorPicker, Input, useTheme2 } from '@grafana/ui';
import React from 'react';

export const ColorPicker = ({
  color,
  onChange,
}: {
  color: string;
  onChange: (color: string) => void;
}) => {
  const theme = useTheme2();

  return (
    <GrafanaColorPicker color={color} onChange={onChange}>
      {({ ref, showColorPicker, hideColorPicker }) => (
        <div style={{ display: 'flex', gap: '0.25rem' }}>
          <div
            onMouseLeave={hideColorPicker}
            ref={ref}
            onClick={showColorPicker}
            style={{
              width: '32px',
              backgroundColor: color,
              borderRadius: theme.shape.borderRadius(1),
              flexShrink: 0,
            }}
          ></div>
          <Input
            value={color}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
          />
        </div>
      )}
    </GrafanaColorPicker>
  );
};
