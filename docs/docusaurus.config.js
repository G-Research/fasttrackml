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

    url: 'https://naskio.github.io',
    baseUrl: '/fasttrackml/',

    organizationName: 'G-Research',
    projectName: 'fasttrackml',

    i18n: {
        defaultLocale: 'en',
        locales: ['en'],
    },

    customFields: {
        email: 'fasttrackml@gr-oss.io',
        newIssueUrl: 'https://github.com/naskio/fasttrackml/issues/new',
    },

    presets: [
        [
            'classic',
            /** @type {import('@docusaurus/preset-classic').Options} */
            ({
                docs: false, // disabling docs
                // docs: {
                //     sidebarPath: require.resolve('./sidebars.js'),
                //     editUrl: 'https://github.com/naskio/fasttrackml/edit/main/docs/',
                //     path: 'content',
                //     exclude: ['example/**'],
                // },
                theme: {
                    customCss: [require.resolve('./src/css/theming.css'), require.resolve('./src/css/announcement-bar.css'), require.resolve('./src/css/global.css')],
                },
            }),
        ],
    ],

    themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
        ({
            image: 'logo/project-logo-text.png', // project's social card
            navbar: {
                title: 'FastTrackML',
                logo: {
                    alt: 'FastTrackML logo',
                    src: 'logo/project-icon.png',
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
                                label: `Found an Issue?`,
                                to: 'https://github.com/G-Research/fasttrackml/issues',
                            },
                            {
                                label: `Create a Pull Request`,
                                to: 'https://github.com/G-Research/fasttrackml/pulls',
                            },
                        ],
                    },
                    {
                        title: 'Community',
                        items: [
                            {
                                label: 'Stack Overflow',
                                href: 'https://stackoverflow.com/questions/tagged/fasttrackml',
                            },
                            {
                                label: 'GitHub Issues',
                                href: 'https://github.com/G-Research/fasttrackml/issues',
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
                    src: 'logo/organization.png',
                    srcDark: 'logo/organization-dark.png',
                    href: 'https://opensource.gresearch.co.uk/',
                },
                copyright: `Copyright © ${new Date().getFullYear()} FastTrackML.`,
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
            metadata: [{name: 'twitter:card', content: 'summary'}],
        }),
};

module.exports = config;
