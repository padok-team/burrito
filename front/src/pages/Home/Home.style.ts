import styled from 'styled-components';
import { Link } from 'react-router-dom';

import { colors, font, getSpacing } from 'stylesheet.ts';
import { ReactComponent as Checkbox } from 'assets/checkbox.svg';
import { ReactComponent as CrossMark } from 'assets/crossmark.svg';
import { ReactComponent as Hourglass } from 'assets/hourglass.svg';

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
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: space-between;
  row-gap: ${getSpacing(2)};
`;

export const Name = styled.div`
  color: ${colors.black};
  ${font.size16};
`;

export const Detail = styled.div`
  color: ${colors.gray1};
  ${font.size14};
`;

export const StatusContainer = styled.div``;

export const Status = styled.div`
  display: flex;
  flex-direction: row;
  align-items: center;
  column-gap: ${getSpacing(1)};
  ${font.size16};
`;

export const CheckboxIcon = styled(Checkbox)`
  height: ${getSpacing(2.5)};
  color: ${colors.green};
`;

export const CrossMarkIcon = styled(CrossMark)`
  height: ${getSpacing(2.5)};
  color: ${colors.red};
`;

export const HourglassIcon = styled(Hourglass)`
  height: ${getSpacing(2.25)};
  color: ${colors.gray1};
`;
