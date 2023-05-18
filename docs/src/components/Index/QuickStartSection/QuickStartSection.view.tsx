import React from 'react';
import clsx from 'clsx';
import MDXContent from '@theme/MDXContent';
import QuickStartMD from "@site/content/quickstart.md";
import styles from './QuickStartSection.module.css';

export default function quickStartSection(): React.JSX.Element {
    return (
        <section>
            <div className={clsx("container")}>
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
