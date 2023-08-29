import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HeroSection from '@site/src/components/Index/HeroSection';
import FeaturesSection from '@site/src/components/Index/FeaturesSection';
import QuickStartSection from '@site/src/components/Index/QuickStartSection';
import ContactUsSection from '@site/src/components/Index/ContactUsSection';


export default function Home(): React.JSX.Element {
    const {siteConfig} = useDocusaurusContext();
    return (
        <Layout
            title="Experiment tracking server focused on scalability"
            description={`${siteConfig.tagline}`}>
            <HeroSection/>
            <main>
                <FeaturesSection/>
                <QuickStartSection/>
                <ContactUsSection/>
            </main>
        </Layout>
    );
}
