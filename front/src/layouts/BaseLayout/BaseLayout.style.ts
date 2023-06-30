import styled from 'styled-components';
import { colors, getSpacing } from 'stylesheet';

export const Container = styled.div`
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: stretch;
`;

export const Header = styled.div`
  padding: 0 ${getSpacing(5)};
  height: ${getSpacing(6)};
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: start;
  column-gap: ${getSpacing(2)};
  border-bottom: 1px solid ${colors.gray0};
`;

export const Title = styled.h1`
  display: flex;
  flex-direction: row;
  align-items: end;
  justify-content: flex-start;
  column-gap: ${getSpacing(1)};
`;

export const Content = styled.div`
  flex-grow: 1;
  overflow-y: scroll;
`;

export const Gutter = styled.div`
  margin: auto;
  max-width: ${getSpacing(192)};
  flex-grow: 1;
  width: 100%;
  box-sizing: border-box;
`;
