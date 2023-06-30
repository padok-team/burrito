import styled from 'styled-components';
import { getSpacing } from 'stylesheet';

export const Container = styled.div`
  padding: ${getSpacing(2)} ${getSpacing(4)};
`;

export const Name = styled.span`
  font-weight: bold;
`;