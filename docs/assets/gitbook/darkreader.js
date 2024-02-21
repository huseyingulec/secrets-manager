document.getElementsByClassName('dark-mode-button')[0].onclick = function(event) {
    event.preventDefault();
    toggleDarkMode();
}

function toggleDarkMode() {
    let enabled = localStorage.getItem('dark-mode');
    const root = document.documentElement;
    const iconDark = document.getElementById('icon-dark');
    const iconLight = document.getElementById('icon-light');

    if (enabled === null) {
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            enable();
        } else {
            disable();
        }
    } else if (enabled === 'true') {
        disable();
    } else {
        enable();
    }

    // Toggle icons and theme
    if (localStorage.getItem('dark-mode') === 'true') {
        iconDark.style.display = 'none';
        iconLight.style.display = 'inline';
        root.classList.remove('light-theme');
        root.classList.add('dark-theme');
    } else {
        iconDark.style.display = 'inline';
        iconLight.style.display = 'none';
        root.classList.remove('dark-theme');
        root.classList.add('light-theme');
    }
}

function enable() {
    localStorage.setItem('dark-mode', 'true');
}

function disable() {
    localStorage.setItem('dark-mode', 'false');
}