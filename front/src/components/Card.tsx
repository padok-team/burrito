
import { GridItem, GridItemProps} from '@chakra-ui/react';

function Card(props : GridItemProps){
    return(
        <GridItem 
            d="flex" 
            borderRadius={10} 
            p={10} 
            boxShadow="md" 
            {...props} 
        />
    )
    

}

export default Card;


