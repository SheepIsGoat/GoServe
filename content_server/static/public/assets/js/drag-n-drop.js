console.log('File loaded');


document.body.addEventListener('htmx:load', function() {
    setupFileDropArea();
});

window.setupComplete = false

function setupFileDropArea() {
    if (window.setupComplete) {
        console.log("FleDropArea is already set up")
        return
    }
    console.log('Setting up FileDropArea');
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

    uploadButton.addEventListener('click', function(){
        fileInput.click();
    });


    fileInput.addEventListener('change', function() {
        handleFiles(this.files);
    });

    // Drag-over and drag-leave visualization
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, unhighlight, false);
    });

    // Handle file drop
    dropArea.addEventListener('drop', handleDrop, false);

    // Helper functions
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    function highlight() {
        dropArea.classList.add('highlight');
    }

    function unhighlight() {
        dropArea.classList.remove('highlight');
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
    console.log("Finished setup")
    window.setupComplete = true
}