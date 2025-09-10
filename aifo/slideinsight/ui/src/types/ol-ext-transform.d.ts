declare module "ol-ext/interaction/Transform" {
  import Interaction from "ol/interaction/Interaction";
  import type Feature from "ol/Feature";

  export interface TransformOptions {
    layers?: (layer: any) => boolean;
    filter?: (feature: Feature) => boolean;
    translate?: boolean;
    scale?: boolean;
    stretch?: boolean;
    rotate?: boolean;
    hitTolerance?: number;
    modifyCenter?: (evt: any) => boolean;
  }

  export default class Transform extends Interaction {
    constructor(options?: TransformOptions);
    setActive(active: boolean): void;
    setSelection(features: Feature[] | []): void;
  }
}
