
import { Heading, Flex, Text} from '@chakra-ui/react';
import Card from './Card';

function LayerCard(props : {id : string, type : string}){
    
    const textFontSize = ["md", "lg", "xl", "xl", "3xl"];

    return(
        <Card flexDirection="column" rowSpan={1} colSpan={1}>
            {props.id && <Heading as="h1" size="2xl">{props.id}</Heading>}
            <Flex flexWrap={["wrap", "wrap", "wrap", "wrap", "nowrap"]} alignItems="center" justifyContent="space-between" marginY="auto">
                <Text fontSize={textFontSize} fontWeight="medium">{props.type}</Text>
            </Flex>
        </Card>
    )
    

}

export default LayerCard;


