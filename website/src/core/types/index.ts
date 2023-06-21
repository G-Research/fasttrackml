import React from "react";

export type FeatureItem = {
    title: string;
    description: React.JSX.Element;
    Svg: React.ComponentType<React.ComponentProps<'svg'>>;
};
