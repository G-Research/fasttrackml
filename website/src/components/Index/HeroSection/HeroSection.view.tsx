import React from "react";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import clsx from "clsx";
import ThemedImage from '@theme/ThemedImage';
import Link from "@docusaurus/Link";
import styles from './HeroSection.module.css';
import useBaseUrl from "@docusaurus/useBaseUrl";

export default function HeroSection(): React.JSX.Element {
    const {siteConfig} = useDocusaurusContext();
    return (
        <header className={clsx('hero', styles.heroBanner)}>
            <div className="container">
                <div className="row row--no-gutters">
                    <div className={clsx("col padding-vert--xl margin-vert--xl", styles.heroHeadingColumn)}>
                        <h1 className="hero__title">{siteConfig.title}</h1>
                        <p className="hero__subtitle">{siteConfig.tagline}</p>
                        <div className={styles.buttons}>
                            <Link
                                className="button button--info button--lg"
                                to="/#quickstart">
                                🚀 Quickstart
                            </Link>
                            <Link
                                className="button button--primary button--lg"
                                to="/#contact-us">
                                {`💬 Talk to Us`}
                            </Link>
                        </div>
                    </div>
                    <div className={clsx("col", styles.heroIconColumn)}>
                        <div className={styles.heroIconContainer}>
                            <div className={styles.heroIconBackground}></div>
                            <ThemedImage
                                className={styles.heroIcon} width={180} height={180}
                                alt="FastTrackML icon"
                                sources={{
                                    light: useBaseUrl('/logo/project-icon-light.svg'),
                                    dark: useBaseUrl('/logo/project-icon-dark.svg'),
                                }}
                            />
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
}
