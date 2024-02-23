console.log('File loaded');

document.body.addEventListener('htmx:load', function() {
    setupFileDropArea();
});

function setupFileDropArea() {
    console.log('Button clicked');
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

    // Open file selector when clicked on the button
    uploadButton.addEventListener('click', function() {
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
        var url = '/upload';
        var formData = new FormData();
        formData.append('file', file);

        // Assuming you're using HTMX for the AJAX call
        htmx.ajax('POST', url, formData, {
            target: '#upload-status'
        });
    }
}
