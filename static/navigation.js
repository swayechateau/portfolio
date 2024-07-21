document.addEventListener('DOMContentLoaded', () => {
    const yAxis = 100; // Adjust this value as needed
    const navigation = document.getElementById('navigation');

    const transitionNavbar = () => {
        if (window.scrollY >= yAxis) {
            navigation.classList.add('bg-main', 'shadow');
            navigation.classList.remove('bg-transparent');
        } else {
            navigation.classList.add('bg-transparent');
            navigation.classList.remove('bg-main', 'shadow');
        }
    };

    window.addEventListener('scroll', transitionNavbar);

    // Clean up event listener if needed
    window.addEventListener('beforeunload', () => {
        window.removeEventListener('scroll', transitionNavbar);
    });
});