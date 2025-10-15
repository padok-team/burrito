import { Layer } from "@/clients/layers/types";

function getLayerType(layer: Layer) {
    if (layer.terraform) {
        return 'terraform';
    } else if (layer.openTofu) {
        return 'opentofu';
    }
    return 'unknown';
}

export { getLayerType };
