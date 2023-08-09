const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

const logoElement = document.createElement('img');
logoElement.className = 'logo';
logoElement.alt = 'full logo';

updateLogo();

darkModeMediaQuery.addListener(updateLogo);

function updateLogo() {
    const logoSrc = darkModeMediaQuery.matches
        ? 'logo/project-logo-dark.svg'
        : 'logo/project-logo-light.svg';
    
    logoElement.src = logoSrc;
    
    const header = document.querySelector('header');
    header.insertBefore(logoElement, header.firstChild);
}
