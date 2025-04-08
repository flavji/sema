document.getElementById('deleteAccountButton').addEventListener('click', function () {
    const confirmed = confirm("Are you sure you want to delete your account? This action cannot be undone.");

    if (!confirmed) return;

    fetch('/api/deleteaccount', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
    })
    .then(response => response.json())
    .then(data => {
        console.log("User has been removed", data);
        document.cookie = "firebaseToken=; Max-Age=0; path=/";
        window.location.href = "/register";
    })
    .catch(error => {
        alert("Error removing user: " + error.message);
        console.error('Error removing user', error);
    });
});
