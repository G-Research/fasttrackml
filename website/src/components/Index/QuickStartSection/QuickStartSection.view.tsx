import React from 'react';
import clsx from 'clsx';
import MDXContent from '@theme/MDXContent';
import QuickStartMD from "./quickstart-section.md";
import styles from './QuickStartSection.module.css';

export default function quickStartSection(): React.JSX.Element {
    return (
        <section className={clsx(styles.section)}>
            <div className={clsx("container padding-vert--md")}>
                <h1 id="quickstart" className={clsx("text--center section__ref")}>Quickstart</h1>
                <div className={clsx(styles.markdownContainer)}>
                    <MDXContent>
                        <QuickStartMD/>
                    </MDXContent>
                </div>
            </div>
        </section>
    );
}
