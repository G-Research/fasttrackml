// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
    title: 'FastTrackML',
    tagline: 'Experiment tracking server focused on speed and scalability',
    favicon: 'favicon.ico',
    onBrokenLinks: 'throw',
    onBrokenMarkdownLinks: 'warn',

    url: 'https://fasttrackml.io',
    baseUrl: '/',

    organizationName: 'G-Research',
    projectName: 'fasttrackml',

    i18n: {
        defaultLocale: 'en',
        locales: ['en'],
    },

    customFields: {
        email: 'fasttrackml@gr-oss.io',
        newIssueUrl: 'https://github.com/G-Research/fasttrackml/issues/new',
    },

    presets: [
        [
            'classic',
            /** @type {import('@docusaurus/preset-classic').Options} */
            ({
                // docs: {
                //     path: '../docs',
                //     sidebarPath: require.resolve('./sidebars.js'),
                //     editUrl: 'https://github.com/G-Research/fasttrackml/edit/main/docs/',
                //     exclude: ['example', 'images'],
                // },
                docs: false, // disabling docs temporarily
                theme: {
                    customCss: [require.resolve('./src/css/theming.css'), require.resolve('./src/css/announcement-bar.css'), require.resolve('./src/css/global.css')],
                },
            }),
        ],
    ],

    themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
        ({
            image: 'images/project-social-preview.png', // project's social card
            navbar: {
                logo: {
                    alt: 'FastTrackML logo',
                    src: 'logo/project-logo-light.svg',
                    srcDark: 'logo/project-logo-dark.svg',
                    width: 140,
                },
                items: [
                    // left
                    {
                        to: '/#quickstart',
                        label: 'Quickstart',
                        position: 'left',
                        activeBaseRegex: `dummy-never-match`,
                    },
                    {
                        to: '/#contact-us',
                        label: 'Contact Us',
                        position: 'left',
                        activeBaseRegex: `dummy-never-match`,
                    },
                    // right
                    {
                        href: 'https://github.com/G-Research/fasttrackml',
                        label: 'GitHub',
                        position: 'right',
                    },
                    {
                        href: 'https://twitter.com/oss_gr',
                        label: 'Twitter',
                        position: 'right',
                    },
                ],
            },
            footer: {
                links: [
                    {
                        title: 'Links',
                        items: [
                            {
                                label: 'Quickstart',
                                to: '/#quickstart',
                            },
                            {
                                label: `Contact Us`,
                                to: '/#contact-us',
                            },
                            {
                                label: `Report an Issue`,
                                to: 'https://github.com/G-Research/fasttrackml/issues',
                            },
                            {
                                label: `Create a Pull Request`,
                                to: 'https://github.com/G-Research/fasttrackml/pulls',
                            },
                        ],
                    },
                    {
                        title: 'More',
                        items: [
                            {
                                label: 'GitHub',
                                href: 'https://github.com/G-Research/fasttrackml',
                            },
                            {
                                label: 'Twitter',
                                href: 'https://twitter.com/oss_gr',
                            },
                            {
                                label: 'G-Research Open-Source',
                                href: 'https://opensource.gresearch.co.uk/',
                            },
                        ],
                    },
                ],
                logo: {
                    alt: 'G-Research Open-Source Software',
                    src: 'logo/organization.svg',
                    srcDark: 'logo/organization-dark.svg',
                    href: 'https://opensource.gresearch.co.uk/',
                },
                copyright: `Copyright © ${new Date().getFullYear()} G-Research`,
            },
            announcementBar: {
                // https://docusaurus.io/docs/api/themes/configuration#announcement-bar
                id: 'announcement-bar--1', // increment on change
                content: `⚠️ FastTrackML is still a work in progress 🚧 and subject to change.`,
                isCloseable: true,
            },
            colorMode: {
                defaultMode: 'light',
                disableSwitch: false,
                respectPrefersColorScheme: true,
            },
            prism: {
                theme: lightCodeTheme,
                darkTheme: darkCodeTheme,
                defaultLanguage: 'bash',
                additionalLanguages: ['python', 'powershell'],
            },
            metadata: [
                {
                    name: 'twitter:card', content: 'summary'
                },
                {
                    name: 'keywords',
                    content: 'machine learning, experiment tracking, mlflow, mlflow tracking server, fasttrackml'
                },
            ],
        }),
};

module.exports = config;
