import React from "react";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import clsx from "clsx";
import ProjectLogoUrl from '@site/static/logo/project-icon.png';
import Link from "@docusaurus/Link";
import styles from './HeroSection.module.css';

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
                                {`💬 Talk with Us`}
                            </Link>
                        </div>
                    </div>
                    <div className={clsx("col", styles.heroIconColumn)}>
                        <div className={styles.heroIconContainer}>
                            <div className={styles.heroIconBackground}></div>
                            <img className={styles.heroIcon} src={ProjectLogoUrl} alt="FastTrackML logo"
                                 width={180} height={180}/>
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
}
