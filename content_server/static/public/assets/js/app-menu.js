// document.addEventListener('DOMContentLoaded', function() {
//     document.querySelectorAll('#navMenu li a').forEach(anchor => {
//         anchor.addEventListener('click', function(event) {
//             // Call toggleIndicator with the clicked <a> element
//             toggleIndicator(this);
//         });
//     });
// });
document.addEventListener('DOMContentLoaded', function() {
    // Event delegation for nav menu
    const navMenu = document.getElementById('navMenu');
    if (navMenu) {
        navMenu.addEventListener('click', function(event) {
            const clickedAnchor = event.target.closest('li a');
            if (clickedAnchor) {
                toggleIndicator(clickedAnchor);
            }
        });
    }

    // Handling dynamic content
    document.body.addEventListener('htmx:load', function() {
        // Reattach or reinitialize event listeners related to dynamic content
        // Example: If you have any specific logic to reinitialize after content load
    });

    // Error page redirection logic
    var errorCodeElem = document.getElementById('errorCode');
    if (errorCodeElem) {
        document.body.addEventListener('click', function(event) {
            const link = event.target.closest('a[hx-get]');
            if (link) {
                event.preventDefault();
                handleHtmxLink(link);
            }
        });
    }
});

function handleHtmxLink(link) {
    // Fetch content using HTMX
    var url = link.getAttribute('hx-get');
    fetch(url)
        .then(response => response.text())
        .then(html => {
            // Update the content area
            var contentArea = document.querySelector('#content-area');
            if (contentArea) {
                contentArea.innerHTML = html;
            }
            // Update URL without reloading the page
            history.pushState({}, '', '/app');
        })
        .catch(error => console.error('Error loading content:', error));
}


function toggleIndicator(clickedAnchor) {
    // Define the classes for active and inactive items
    const activeClass = "inline-flex items-center w-full text-sm font-semibold text-gray-800 transition-colors duration-150 hover:text-gray-800 dark:hover:text-gray-200 dark:text-gray-100";
    const inactiveClass = "inline-flex items-center w-full text-sm font-semibold transition-colors duration-150 hover:text-gray-800 dark:hover:text-gray-200";

    // Hide all indicators and reset the class of all <a> tags
    document.querySelectorAll('#navMenu li').forEach(item => {
        const indicator = item.querySelector('.indicator');
        if (indicator) {
            indicator.classList.add('hidden');
        }
        const anchor = item.querySelector('a');
        if (anchor) {
            anchor.className = inactiveClass;
        }
    });

    const listItem = clickedAnchor.closest('li');
    if (listItem) {
        const indicatorSpan = listItem.querySelector('.indicator');
        if (indicatorSpan) {
            indicatorSpan.classList.remove('hidden');
        }
        clickedAnchor.className = activeClass;
    }

}

// document.addEventListener('DOMContentLoaded', function() {
//     // redirects error pages to /app while maintaining link integrity
//     // especially useful for 404 pages not served at /app
//     var errorCodeElem = document.getElementById('errorCode');

//     if (errorCodeElem) {
//         var links = document.querySelectorAll('a[hx-get]');
//         links.forEach(function(link) {
//             link.addEventListener('click', function(event) {
//                 event.preventDefault();

//                 // Fetch content using HTMX
//                 var url = this.getAttribute('hx-get');
//                 fetch(url)
//                     .then(response => response.text())
//                     .then(html => {
//                         // Update the content area
//                         var contentArea = document.querySelector('#content-area');
//                         if (contentArea) {
//                             contentArea.innerHTML = html;
//                         }

//                         // Update URL without reloading the page
//                         history.pushState({}, '', '/app');
//                     })
//                     .catch(error => console.error('Error loading content:', error));
//             });
//         });
//     }
// });