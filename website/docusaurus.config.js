// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

module.exports = async function configCreatorAsync() {
    const lightCodeTheme = require('prism-react-renderer/themes/github');
    const darkCodeTheme = require('prism-react-renderer/themes/dracula');

    // Get latest release version from GitHub
    const { Octokit } = require('@octokit/rest');
    const octokit = new Octokit();
    const { data } = await octokit.request('GET /repos/{owner}/{repo}/releases/latest',
        {
            owner: 'G-Research',
            repo: 'fasttrackml'
        }
    );
    const releaseVersion = data.tag_name.substring(1);
    const releaseUrl = data.html_url;

    /** @type {import('@docusaurus/types').Config} */
    return {
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
                    docs: false, // disabling docs
                    theme: {
                        customCss: [require.resolve('./src/css/theming.css'), require.resolve('./src/css/announcement-bar.css'), require.resolve('./src/css/global.css')],
                    },
                    gtag: {
                        trackingID: 'G-2YZLJEB3PY',
                        anonymizeIP: true,
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
                    id: 'announcement-bar',
                    content: `FastTrackML ${releaseVersion} has been <a href="${releaseUrl}" target="_blank">released</a>!`,
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
};