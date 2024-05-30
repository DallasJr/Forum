function isValidUsername(username) {
    const usernameRegex = /^[a-zA-Z0-9_.-]+$/;
    return usernameRegex.test(username) && username.length >= 3 && username.length <= 16;
}
function isValidNameOrSurname(nameOrSurname) {
    const nameOrSurnamePattern = /^[a-zA-Z-]+$/;
    return nameOrSurnamePattern.test(nameOrSurname);
}

function checkPassword(password) {
    let password8min = document.getElementById("password8min");
    let passwordNumber = document.getElementById("passwordNumber");
    let passwordSpecialChar = document.getElementById("passwordSpecialChar");
    //let passwordNoSpaces = document.getElementById("passwordNoSpaces");
    let min8
    if (password.length >= 8) {
        password8min.textContent = "✓ At least 8 characters";
        password8min.style.color = "green";
        min8 = true
    } else {
        password8min.textContent = "✗ At least 8 characters";
        password8min.style.color = "red";
        min8 = false
    }
    let number
    if (/\d/.test(password)) {
        passwordNumber.textContent = "✓ Contains at least one number";
        passwordNumber.style.color = "green";
        number = true
    } else {
        passwordNumber.textContent = "✗ Contains at least one number";
        passwordNumber.style.color = "red";
        number = false
    }

    let specialChar
    if (/[^a-zA-Z0-9]/.test(password)) {
        passwordSpecialChar.textContent = "✓ Contains at least one special character";
        passwordSpecialChar.style.color = "green";
        specialChar = true
    } else {
        passwordSpecialChar.textContent = "✗ Contains at least one special character";
        passwordSpecialChar.style.color = "red";
        specialChar = false
    }
    /*let noSpaces
    if (/\s/.test(password)) {
        passwordNoSpaces.textContent = "✗ Password must not contain spaces";
        passwordNoSpaces.style.color = "red";
        noSpaces = false
    } else {
        passwordNoSpaces.textContent = "";
        passwordNoSpaces.style.color = "green";
        noSpaces = true
    }*/
    return min8 && number && specialChar/* && noSpaces*/;
}