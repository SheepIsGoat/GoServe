document.body.removeEventListener('htmx:load', dropDownManager);

document.body.addEventListener('htmx:load', dropDownManager);

function dropDownManager() {
    const menuButton = document.getElementById('menu-button');
    const dropdownMenu = document.getElementById('dropdown-menu');

    if (menuButton && dropdownMenu) {
        menuButton.removeEventListener('click', toggleHidden);
        menuButton.addEventListener('click', toggleHidden);

        function toggleHidden() {
            dropdownMenu.classList.toggle('hidden');
        }

        window.removeEventListener('click', closeDropdown);
        window.addEventListener('click', closeDropdown);

        function closeDropdown(e) {
            if (!menuButton.contains(e.target) && !dropdownMenu.contains(e.target)) {
                dropdownMenu.classList.add('hidden');
            }
            e.stopPropogation();
        }
    }
}