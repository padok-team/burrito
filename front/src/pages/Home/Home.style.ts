import styled from 'styled-components';
import { colors, font, getSpacing } from 'stylesheet.ts';
import { Link } from 'react-router-dom';

export const Container = styled.div`
  padding: ${getSpacing(2)} ${getSpacing(4)};
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: ${getSpacing(2)};
`;

export const Card = styled(Link)`
  padding: ${getSpacing(2)};
  border: 1px solid ${colors.gray0};
  border-radius: 4px;
  cursor: pointer;
`;

export const Name = styled.div`
  color: ${colors.black};
  ${font.size16};
`;

export const Detail = styled.div`
  color: ${colors.gray1};
  ${font.size14};
`;
