// /static/js/chat.js

document.addEventListener("DOMContentLoaded", () => {
  const messages = document.getElementById('messages');
  const input = document.getElementById('input');
  const messageInput = document.getElementById('message');
  const status = document.getElementById('status');

  // Get username from the DOM safely
  let currentUsername = 'Guest';
  try {
    const usernameElement = document.querySelector('.username');
    if (usernameElement && usernameElement.textContent) {
      currentUsername = usernameElement.textContent.trim();
    }
  } catch (error) {
    console.log('Using default username');
  }

  // Generate a unique ID for this user session
  const userId = 'user_' + Math.random().toString(36).substr(2, 9);

  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = wsProtocol + '//' + window.location.host + '/room';

  const conn = new WebSocket(wsUrl);

  conn.onopen = function () {
    status.textContent = 'Connected';
    status.className = 'status connected';
  };

  conn.onclose = function () {
    status.textContent = 'Disconnected';
    status.className = 'status disconnected';
  };

  conn.onerror = function () {
    status.textContent = 'Connection Error';
    status.className = 'status disconnected';
  };

  conn.onmessage = function (e) {
    try {
      const msg = JSON.parse(e.data);
      const messageDiv = document.createElement('div');

      if (msg.userId === userId) {
        messageDiv.className = 'message sent';
        messageDiv.innerHTML = `
          <div class="message-content">${msg.message}</div>
        `;
      } else {
        messageDiv.className = 'message received';
        messageDiv.innerHTML = `
          <div class="message-sender">${msg.name || 'User'}</div>
          <div class="message-content">${msg.message}</div>
        `;
      }

      messages.appendChild(messageDiv);
      messages.scrollTop = messages.scrollHeight;

    } catch {
      const message = document.createElement('div');
      message.className = 'message received';
      message.textContent = e.data;
      messages.appendChild(message);
      messages.scrollTop = messages.scrollHeight;
    }
  };

  input.addEventListener('submit', function (e) {
    e.preventDefault();
    const messageText = messageInput.value.trim();

    if (messageText) {
      const messageData = {
        type: 'message',
        message: messageText,
        userId: userId,
        name: currentUsername,
        timestamp: new Date().toISOString()
      };

      conn.send(JSON.stringify(messageData));
      messageInput.value = '';
      messageInput.focus();
    }
  });

  messageInput.focus();
});
