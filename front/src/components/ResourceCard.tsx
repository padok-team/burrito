import { Heading, Flex, Text } from '@chakra-ui/react';
import Card from './Card';
import { Resource } from 'client/layers/type';
import { Name, DependenciesList } from './Card.style';

interface Props {
  resource: Resource;
}

const ResourceCard: React.FC<Props> = ({ resource }) => {
  const textFontSize = ['s'];

  return (
    <Card flexDirection="column" rowSpan={1} colSpan={1}>
      {resource.address && (
        <Heading as="h1" size="md">
          {resource.address}
        </Heading>
      )}
      <Flex
        flexWrap={['wrap', 'wrap', 'wrap', 'wrap', 'nowrap']}
        alignItems="center"
        justifyContent="space-between"
        marginY="auto"
      >
        <div>
          <div>
            <Name>Type: </Name>
            <span>{resource.type}</span>
          </div>
          <div>
            <Name>Status: </Name>
            <span>{resource.status}</span>
          </div>
          <div>
            <Name>depends_on: </Name>
            <DependenciesList>
              {resource.depends_on?.map((dependency) => (
                <div>{dependency}</div>
              ))}
            </DependenciesList>
          </div>
        </div>
      </Flex>
    </Card>
  );
};

export default ResourceCard;
