import React from 'react';
import clsx from 'clsx';
import styles from './FeaturesSection.module.css';
import {FeatureItem} from "@site/src/core/types";

const FeatureList: FeatureItem[] = [
    {
        title: 'Blazing Fast',
        Svg: require('@site/static/images/fast.svg').default,
        description: (
            <>
                FastTrackML is a rewrite of the MLFlow tracking server with a focus on
                performance and scalability.
            </>
        ),
    },
    {
        title: 'Easy to Use',
        Svg: require('@site/static/images/easy.svg').default,
        description: (
            <>
                FastTrackML is designed to be easily installed and used to get your experiments tracked quickly.
                Use the Modern UI alternative for a seamless experience.
            </>
        ),
    },
    {
        title: 'Drop-in Replacement',
        Svg: require('@site/static/images/drop-in.svg').default,
        description: (
            <>
                Use the Classic UI to get the same experience as MLFlow's tracking server. But yet much faster than
                MLFlow's.
            </>
        ),
    },
];


function Feature({Svg, title, description}: FeatureItem) {
    return (
        <div className={clsx('col col--4')}>
            <div className="card">
                <div className="card__image text--center padding-top--lg">
                    <Svg className={styles.featureSvg} role="img"/>
                </div>
                <div className="card__body text--center">
                    <h3>{title}</h3>
                    <p>{description}</p>
                </div>
            </div>
        </div>
    );
}

export default function FeaturesSection(): React.JSX.Element {
    return (
        <section>
            <div className="container padding-top--md">
                <h1 id="highlights" className={clsx("text--center section__ref")}>Highlights</h1>
                <div className={clsx("row padding-top--md", styles.features)}>
                    {FeatureList.map((props, idx) => (
                        <Feature key={idx} {...props} />
                    ))}
                </div>
            </div>
        </section>
    );
}
