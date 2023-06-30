import { css } from 'styled-components';

const SPACING_UNIT = 8;
const MEASUREMENT_UNIT = 'px';

const getPixels = (multiplier: number): number => multiplier * SPACING_UNIT;

export const getSpacing = (multiplier: number): string =>
  `${getPixels(multiplier)}${MEASUREMENT_UNIT}`;

export const colors = {
  black: '#2A2A2A',
  gray1: '#9A9A9A',
  gray0: '#c5c5c5',
};

export const font = {
  size14: css`
    font-size: 14px;
    line-height: 20px;
  `,
  size16: css`
    font-size: 16px;
    line-height: 24px;
  `,
  size20: css`
    font-size: 20px;
    line-height: 28px;
  `,
  size24: css`
    font-size: 24px;
    line-height: 28px;
  `,
};
