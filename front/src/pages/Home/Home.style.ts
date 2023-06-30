import styled from 'styled-components';
import { colors, getSpacing } from 'stylesheet.ts';

export const Container = styled.div`
  padding: ${getSpacing(2)} ${getSpacing(4)};
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: ${getSpacing(2)};
`;

export const Card = styled.div`
  padding: ${getSpacing(2)};
  border: 1px solid ${colors.gray0};
  border-radius: 4px;
`;
