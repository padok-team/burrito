
import { GridItem, GridItemProps} from '@chakra-ui/react';

function Card(props : GridItemProps){
    return(
        <GridItem 
            d="flex" 
            borderRadius={10} 
            p={5} 
            boxShadow="md" 
            {...props} 
        />
    )
    

}

export default Card;


