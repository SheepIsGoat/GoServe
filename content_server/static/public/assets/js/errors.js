function loadError(event) {
    let status = event.detail.xhr.status;
    let errorMessage = "An error occurred.";
    switch (status) {
        case 404:
            errorMessage = "Resource not found.";
            break;
        case 500:
            errorMessage = "Internal server error.";
            break;
        // Handle other statuses as needed
    }
    // Display error (e.g., modal, toast, etc.)
    triggerAlert(errorMessage);
}

function triggerAlert(message) {
    let alertBox = document.querySelector('[x-data]');
    Alpine.data(alertBox).__x.$data.showAlert = true;
    Alpine.data(alertBox).__x.$data.message = message;
}