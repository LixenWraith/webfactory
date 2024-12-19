document.querySelectorAll('.link').forEach(link => {
    link.addEventListener('mouseover', () => link.style.textDecoration = 'none');
});