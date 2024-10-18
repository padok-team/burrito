import React, { useState, useEffect, useRef } from 'react';
import Checkbox from '@/components/core/Checkbox';
import Button from '@/components/core/Button';
import Tag from '@/components/widgets/Tag';
import { Layer } from '@/clients/layers/types';

// Define the props for the LayerChecklist component
interface LayerChecklistProps {
    layers: Layer[];
    variant?: 'light' | 'dark'; // Optional: To pass variant to Checkbox
    onSelectionChange?: (selectedLayers: { name: string; namespace: string }[]) => void; // Updated callback prop
}

const LayerChecklist: React.FC<LayerChecklistProps> = ({
    layers,
    variant = 'light',
    onSelectionChange,
}) => {
    // State to keep track of selected layers using unique keys
    const [selectedLayers, setSelectedLayers] = useState<{ name: string; namespace: string }[]>([]);
    const selectAllRef = useRef<HTMLInputElement>(null);

    // Function to generate a unique key for each layer
    const getLayerKey = (layer: Layer): string => `${layer.namespace}-${layer.name}`;

    // Update the indeterminate state based on selection
    useEffect(() => {
        if (selectAllRef.current) {
            const isIndeterminate =
                selectedLayers.length > 0 && selectedLayers.length < layers.length;
            selectAllRef.current.indeterminate = isIndeterminate;
        }
    }, [selectedLayers, layers.length]);

    // Handler for individual layer checkbox toggle
    const handleToggle = (layer: Layer) => {
        setSelectedLayers((prevSelected) =>
            prevSelected.some((selectedLayer) => selectedLayer.name === layer.name && selectedLayer.namespace === layer.namespace)
                ? prevSelected.filter((selectedLayer) => selectedLayer.name !== layer.name || selectedLayer.namespace !== layer.namespace)
                : [...prevSelected, { name: layer.name, namespace: layer.namespace }]
        );
    };

    // Handler to select all layers
    const handleSelectAll = () => {
        setSelectedLayers(layers.map(layer => ({ name: layer.name, namespace: layer.namespace })));
    };

    // Handler to unselect all layers
    const handleUnselectAll = () => {
        setSelectedLayers([]);
    };

    useEffect(() => {
        if (onSelectionChange) {
            onSelectionChange(selectedLayers);
        }
    }, [selectedLayers, onSelectionChange]);

    return (
        <div className="max-w-md mx-auto p-4 bg-white shadow-md rounded-md">
            <div className="flex justify-start items-center mb-4 space-x-2">
                <Button
                    variant={"tertiary"}
                    theme={variant}
                    className='text-sm px-0'
                    onClick={handleSelectAll}>
                    Select All
                </Button>
                <Button
                    variant={"tertiary"}
                    theme={variant}
                    className='text-sm'
                    onClick={handleUnselectAll}>
                    Unselect All
                </Button>
            </div>
            <ul className="space-y-2">
                {layers.map((layer) => {
                    const key = getLayerKey(layer);
                    return (
                        <li key={key}>
                            <div className="flex justify-between">
                            <Checkbox
                                label={`${layer.namespace}/${layer.name}`}
                                checked={selectedLayers.some(selectedLayer => selectedLayer.name === layer.name && selectedLayer.namespace === layer.namespace)}
                                onChange={() => handleToggle(layer)}
                                variant={variant}
                            />
                            <Tag variant={layer.state}/>
                            </div>
                        </li>
                    );
                })}
            </ul>
        </div>
    );
};

export default LayerChecklist;
