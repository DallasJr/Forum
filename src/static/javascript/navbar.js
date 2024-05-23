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

document.addEventListener("DOMContentLoaded", function() {
    function getCookie(name) {
        let matches = document.cookie.match(new RegExp(
            "(?:^|; )" + name.replace(/([.$?*|{}()[]\\\/+^])/g, '\\$1') + "=([^;]*)"
        ));
        return matches ? decodeURIComponent(matches[1]) : undefined;
    }

    const sessionID = getCookie("sessionID");

    const loginElement = document.getElementById("login");
    const registerElement = document.getElementById("register");
    const loginmElement = document.getElementById("login-m");
    const registermElement = document.getElementById("register-m");
    const logoutElement = document.getElementById("logout");

    if (sessionID) {
        if (loginElement) loginElement.style.display = 'none';
        if (loginmElement) loginmElement.style.display = 'none';
        if (registerElement) registerElement.style.display = 'none';
        if (registermElement) registermElement.style.display = 'none';
        if (logoutElement) logoutElement.style.display = 'block';
    } else {
        if (loginElement) loginElement.style.display = 'block';
        if (loginmElement) loginmElement.style.display = 'block';
        if (registerElement) registerElement.style.display = 'block';
        if (registermElement) registermElement.style.display = 'block';
        if (logoutElement) logoutElement.style.display = 'none';
    }
});