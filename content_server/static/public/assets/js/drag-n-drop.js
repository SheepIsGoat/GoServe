
let isSetupComplete = false;
if (!isSetupComplete){
    document.body.addEventListener('htmx:load', setupFileDropArea);
    isSetupComplete = true;
}

function triggerFileInputClick() {
    var fileInput = document.getElementById('file-input');
    if (fileInput) {
        fileInput.click();
    }
}

function fileHandler() {
    handleFiles(this.files);
};

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

function highlight(e) {
    e.currentTarget.classList.add('highlight');
}

function unhighlight(e) {
    e.currentTarget.classList.remove('highlight');
}

function handleDrop(e) {
    var dt = e.dataTransfer;
    handleFiles(dt.files);
}

function handleFiles(files) {
    ([...files]).forEach(uploadFile);
}

function uploadFile(file) {
    var url = '/app/upload';
    var formData = new FormData();
    formData.append('file', file);

    console.log("Uploading file...")
    // Using Fetch API as an alternative to HTMX for the AJAX call
    fetch(url, {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return response.text();
    }).then(html => {
        // Update the target element with the response
        var uploadStatus = document.querySelector('#upload-status');
        uploadStatus.innerHTML = html;
        setTimeout(function() { uploadStatus.innerHTML=""; }, 5000);
    })
    .catch(error => console.error(error));
}

function setupFileDropArea() {
    var dropArea = document.getElementById('drop-area');
    if (!dropArea) {
        console.log("No drop area found")
        return
    }
    var fileInput = document.getElementById('file-input');
    if (!fileInput) {
        console.log("No file-input found")
        return
    }
    var uploadButton = dropArea.querySelector('button');

    uploadButton.removeEventListener('click', triggerFileInputClick);
    uploadButton.addEventListener('click', triggerFileInputClick);

    fileInput.removeEventListener('change', fileHandler);
    fileInput.addEventListener('change', fileHandler);
    

    // Drag-over and drag-leave visualization
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.removeEventListener(eventName, preventDefaults, false);
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.removeEventListener(eventName, highlight, false);
        dropArea.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.removeEventListener(eventName, unhighlight, false);
        dropArea.addEventListener(eventName, unhighlight, false);
    });

    // Handle file drop
    dropArea.removeEventListener('drop', handleDrop, false);
    dropArea.addEventListener('drop', handleDrop, false);
    
    console.log("Finished setup")
}