import React from 'react';
import clsx from 'clsx';
import styles from './CommunitySection.module.css';
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";

export default function CommunitySection(): React.JSX.Element {
    const {
        siteConfig: {customFields}
    } = useDocusaurusContext();
    const slackWorkspaceInviteUrl = customFields?.slackWorkspaceInviteUrl as string;
    const slackChannelUrl = customFields?.slackChannelUrl as string;

    return (
        <section className={clsx(styles.section)}>
            <div className="container padding-bottom--xl text--center">
                <h1 id="community" className={clsx("section__ref")}>Community</h1>
                <p>We invite you to join our community on Slack:</p>
                <p>
                    - If you haven't joined the <strong>MLOps.community</strong> workspace yet, please sign
                    up <a
                    href={slackWorkspaceInviteUrl} target="_blank">here</a>.
                </p>
                <p>
                    - Once you're a member, feel free to join us in the{` `}
                    <strong><a href={slackChannelUrl} target="_blank">#fasttrackml</a></strong> channel.
                </p>
                <p>We look forward to seeing you there!</p>
            </div>
        </section>
    );
}
