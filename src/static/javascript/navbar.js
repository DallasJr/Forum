function toggleMenu() {
    var x = document.getElementById("mobile-links");
    if (x.style.display === "flex") {
        x.style.display = "none";
    } else {
        x.style.display = "flex";
    }
}
window.addEventListener('resize', function(event){
    var navLinks = document.getElementById("nav-links");
    var mobileLinks = document.getElementById("mobile-links");
    if (window.innerWidth > 768) {
        navLinks.style.display = "flex";
        mobileLinks.style.display = "none";
    } else {
        navLinks.style.display = "none";
    }
});

