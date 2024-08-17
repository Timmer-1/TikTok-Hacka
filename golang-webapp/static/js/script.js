document.getElementById('contentForm').addEventListener('submit', function (e) {
    e.preventDefault();

    const contentInput = document.getElementById('contentInput');
    const content = contentInput.value.trim();

    if (content) {
        addContentToList(content);
        contentInput.value = '';
    }

    // Disable the submit button for 2 seconds to prevent multiple submissions
    const submitButton = document.getElementById("submit-button");
    submitButton.disabled = true;
    setTimeout(function () {
        submitButton.disabled = false;
    }, 2000);
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



let socket = new WebSocket("ws://localhost:8080/notifications");

socket.onopen = function () {
    console.log("WebSocket connection established.");
};

socket.onmessage = function (event) {
    console.log("WebSocket message received:", event.data);
    let newMessage = event.data;
    showNotification("New content posted: " + newMessage);
};

socket.onclose = function (event) {
    if (event.wasClean) {
        console.log(`WebSocket connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        console.error('WebSocket connection closed abruptly.');
    }
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

    // Hide the notification after 3 seconds
    setTimeout(function () {
        notificationModal.classList.remove("show");
    }, 3000);
}

function deleteMessage(event, id) {
    event.stopPropagation();  // Prevent the click event from bubbling up to the li element

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

socket.onopen = function () {
    console.log("WebSocket connection established.");
};

socket.onmessage = function (event) {
    console.log("WebSocket message received:", event.data);
    let newMessage = event.data;
    showNotification("New content posted: " + newMessage);
};

socket.onclose = function (event) {
    if (event.wasClean) {
        console.log(`WebSocket connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        console.error('WebSocket connection closed abruptly.');
    }
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

    // Hide the notification after 3 seconds
    setTimeout(function () {
        notificationModal.classList.remove("show");
    }, 3000);
}

function addContentToList(content) {
    const contentList = document.getElementById('contentList');
    const li = document.createElement('li');
    li.textContent = content;
    contentList.appendChild(li);
}

function favoriteContent(msgID) {
    fetch('/submitRecommend', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 'msg_id': msgID }),  // Corrected this line
    })
        .then(response => response.json())
        .then(data => {
            console.log('Success:', data);
        })
        .catch((error) => {
            console.error('Error:', error);
        });
}