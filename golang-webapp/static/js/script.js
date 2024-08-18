document.addEventListener('DOMContentLoaded', function () {
    let socket = new WebSocket("ws://localhost:8080/notifications");

    socket.onopen = function () {
        console.log("WebSocket connection established.");
    };

    socket.onmessage = function (event) {
        console.log("WebSocket message received:", event.data);
        showNotification("New content posted: " + event.data);
    };

    socket.onclose = function (event) {
        if (event.wasClean) {
            console.log(`WebSocket connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
            console.error('WebSocket connection closed abruptly.');
        }
    }

    // Add conditional class based on content length
    document.querySelectorAll('#contentList li').forEach(function (item) {
        const contentText = item.querySelector('.content-text').textContent;

        // Assuming "long text" is anything longer than 100 characters
        if (contentText.length > 100) {
            item.classList.add('long-text');
        }
    });

    document.getElementById('contentForm').addEventListener('submit', function (e) {
        e.preventDefault();

        const contentInput = document.getElementById('contentInput');
        const content = contentInput.value.trim();

        if (content) {
            this.submit(); // Only submit if the content is not empty
        } else {
            alert("Content cannot be empty");
        }
    });
});
function deleteMessage(event, id) {
    event.stopPropagation();  // Prevent the click event from bubbling up to the li element
    console.log(`Delete button clicked for message ID: ${id}`);  // Debugging log

    if (confirm("Are you sure you want to delete this message?")) {
        fetch(`/delete?id=${id}`, {
            method: 'POST'
        })
            .then(response => {
                if (response.ok) {
                    // Remove the content item from the DOM
                    document.getElementById(`message-${id}`).remove();
                } else {
                    console.error("Failed to delete message");
                }
            })
            .catch(error => console.error('Error:', error));
    }
}
socket.onerror = function (error) {
    console.error("WebSocket error:", error);
};

socket.onerror = function (error) {
    console.error("WebSocket error:", error);
};

window.addEventListener("beforeunload", function () {
    socket.close();
});

// Function to show a notification modal
function showNotification(message) {
    const notificationModal = document.getElementById("notification-modal");
    notificationModal.textContent = message;
    notificationModal.classList.add("show");

    // Clear any existing timeouts to avoid flickering if notifications come in quick succession
    clearTimeout(notificationModal.hideTimeout);

    // Hide the notification after 3 seconds
    notificationModal.hideTimeout = setTimeout(function () {
        notificationModal.classList.remove("show");
    }, 3000);
}

function addContentToList(content) {
    const contentList = document.getElementById('contentList');
    const trimmedContent = content.trim(); // Trim whitespace
    const li = document.createElement('li');
    li.innerHTML = `<div class="content-text">${trimmedContent}</div>
                        <div class="button-container">
                            <button class="favorite" onclick="favoriteContent('${li.id}', '${trimmedContent}')">Favorite</button>
                            <button class="delete" onclick="deleteMessage(event, '${li.id}')">Delete</button>
                        </div>`;
    contentList.appendChild(li);
}


function favoriteContent(msgID, content) {
    console.log("favoriteContent function called with:", msgID, content);

    const favoriteList = document.getElementById('favoriteList');

    // Create a new list item for the favorite
    const li = document.createElement('li');
    li.id = `favorite-${msgID}`;
    li.innerHTML = `
        <div class="content-text">${content}</div>
        <div class="button-container">
            <button class="delete" onclick="deleteFavorite(event, '${msgID}')">Remove from Favorites</button>
        </div>`;
    favoriteList.appendChild(li);

    // Send favorite request to the server
    fetch('/submitRecommend', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 'msg_id': msgID }),
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => { throw new Error(text); });
            }
            return response.json();
        })
        .then(data => {
            console.log('Success:', data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}



function deleteFavorite(event, id) {
    event.stopPropagation();  // Prevent the click event from bubbling up to the li element

    if (confirm("Are you sure you want to remove this message from favorites?")) {
        fetch(`/deleteFavorite?id=${id}`, {
            method: 'POST'
        })
            .then(response => {
                if (response.ok) {
                    // Remove the favorite item from the DOM
                    document.getElementById(`favorite-${id}`).remove();
                } else {
                    console.error("Failed to remove favorite message");
                }
            })
            .catch(error => console.error('Error:', error));
    }
}