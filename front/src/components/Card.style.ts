import styled from 'styled-components';
import { colors, getSpacing } from 'stylesheet';

export const Container = styled.div`
  padding: ${getSpacing(2)} ${getSpacing(4)};
`;

export const Name = styled.span`
  font-weight: bold;
`;

export const DependenciesList = styled.div`
  padding: ${getSpacing(0)} ${getSpacing(2)};
`;
