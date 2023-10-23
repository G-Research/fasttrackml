import React from 'react';
import clsx from 'clsx';
import {useForm, SubmitHandler} from "react-hook-form";
import styles from './ContactUsSection.module.css';
import {openNewIssue, openNewEmail, openInNewTab} from "@site/src/core/utils";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";


type Inputs = {
    name: string,
    company: string,
    subject: string,
    message: string,
};

const getMessage = (name: string, company: string, message: string): string => {
    let result = "";
    if (name && company) {
        result += `${name} from ${company},\n`;
    }
    if (message) {
        result += `${message}\n`;
    }
    return result;
}

export default function ContactUsSection(): React.JSX.Element {
    const {register, handleSubmit, formState: {isValid}} = useForm<Inputs>();
    const {
        siteConfig: {customFields}
    } = useDocusaurusContext();
    const email = customFields?.email;
    const slackInviteUrl = customFields?.slackInviteUrl

    const onSendEmail: SubmitHandler<Inputs> = data => {
        openNewEmail(email as string, data.subject, getMessage(data.name, data.company, data.message));
    };

    const onJoinSlack: SubmitHandler<Inputs> = data => {
        openInNewTab(slackInviteUrl as string);
    };

    return <section>
        <div className="container padding-bottom--xl text--center">
            <h1 id="contact-us" className={clsx("section__ref")}>Contact Us</h1>
            <p>We would love to hear from you! FastTrackML is a brand new project and any contribution would make a difference!</p>
            <form onSubmit={handleSubmit(onSendEmail)}>
                <div className='row'>
                    <div className='col col--6'>
                        <label htmlFor="name">Your Name</label>
                        <input id="name" placeholder="Your Name"
                               className="button--outline button--secondary button--lg"
                               {...register("name")}
                        />
                    </div>
                    <div className="col col--6">
                        <label htmlFor="company">Company Name</label>
                        <input id="company" placeholder="Company Name"
                               className="button--outline button--secondary button--lg"
                               {...register("company")}
                        />
                    </div>
                </div>
                <div className="row">
                    <div className="col">
                        <label htmlFor="subject">Subject</label>
                        <input id="subject" placeholder="Subject"
                               className="button--outline button--secondary button--lg"
                               {...register("subject")}
                        />
                    </div>
                </div>
                <div className="row">
                    <div className="col">
                        <label htmlFor="message">Message</label>
                        <textarea id="message"
                                  placeholder={`Questions, suggestions, or feedback. Any contribution would make a difference!`}
                                  className="button--outline button--secondary button--lg" rows={4}
                                  {...register("message")}
                        />
                    </div>
                </div>
                <div className="row margin-vert--md">
                    <div className={clsx("col", styles.buttons)}>
                        <button type="button" className="button button--outline button--primary button--lg"
                                title={email as string}
                                disabled={!isValid} onClick={handleSubmit(onSendEmail)}>
                            📨 Submit
                        </button>
                    </div>
                </div>
            </form>
	    <p>Or, join the <b>#fasttrackml</b> channel on the <b>MLOps.community</b> Slack!</p>
            <div className="row margin-vert--md">
		<div className={clsx("col", styles.buttons)}>
		    <button type="submit" className="button button--primary button--lg"
			    title={slackInviteUrl as string}
			    disabled={!isValid} onClick={handleSubmit(onJoinSlack)}>
			💬 Join our Slack
		    </button>
		</div>		
	    </div>
        </div>
    </section>;
}
