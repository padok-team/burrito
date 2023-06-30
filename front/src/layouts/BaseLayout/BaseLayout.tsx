import React, { ReactNode } from 'react';

import { Container, Header, Title, Content } from './BaseLayout.style';
import { FormattedMessage } from 'react-intl';

interface Props {
  children: ReactNode;
}

const BaseLayout: React.FC<Props> = ({ children }) => {
  return (
    <Container>
      <Header>
        <Title>
          <FormattedMessage id="title" />
        </Title>
      </Header>
      <Content>{children}</Content>
    </Container>
  );
};

export default BaseLayout;
