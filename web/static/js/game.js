/* =====================================================
       CANVAS
    ===================================================== */

const canvas = document.getElementById("canvas");

const ctx = canvas.getContext("2d");

let drawing = false;
let lastX = 0;
let lastY = 0;

let color = "#25232b";
let size = 5;
let erasing = false;

function resizeCanvas() {
  const rect = canvas.getBoundingClientRect();

  const oldCanvas = document.createElement("canvas");

  oldCanvas.width = canvas.width;
  oldCanvas.height = canvas.height;

  if (canvas.width && canvas.height) {
    oldCanvas.getContext("2d").drawImage(canvas, 0, 0);
  }

  canvas.width = rect.width;
  canvas.height = rect.height;

  ctx.fillStyle = "#ffffff";

  ctx.fillRect(0, 0, canvas.width, canvas.height);

  if (oldCanvas.width) {
    ctx.drawImage(
      oldCanvas,
      0,
      0,
      oldCanvas.width,
      oldCanvas.height,
      0,
      0,
      canvas.width,
      canvas.height,
    );
  }
}

function position(event) {
  const rect = canvas.getBoundingClientRect();

  if (event.touches) {
    return {
      x: event.touches[0].clientX - rect.left,

      y: event.touches[0].clientY - rect.top,
    };
  }

  return {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top,
  };
}

function startDrawing(event) {
  event.preventDefault();

  drawing = true;

  const p = position(event);

  lastX = p.x;
  lastY = p.y;

  draw(event);
}

function draw(event) {
  if (!drawing) return;

  event.preventDefault();

  const p = position(event);

  ctx.beginPath();

  ctx.moveTo(lastX, lastY);

  ctx.lineTo(p.x, p.y);

  ctx.strokeStyle = erasing ? "#ffffff" : color;

  ctx.lineWidth = size;

  ctx.lineCap = "round";
  ctx.lineJoin = "round";

  ctx.stroke();

  lastX = p.x;
  lastY = p.y;
}

function stopDrawing() {
  drawing = false;

  ctx.beginPath();
}

canvas.addEventListener("mousedown", startDrawing);

canvas.addEventListener("mousemove", draw);

canvas.addEventListener("mouseup", stopDrawing);

canvas.addEventListener("mouseleave", stopDrawing);

canvas.addEventListener("touchstart", startDrawing, { passive: false });

canvas.addEventListener("touchmove", draw, { passive: false });

canvas.addEventListener("touchend", stopDrawing);

window.addEventListener("resize", resizeCanvas);

/* =====================================================
       TOOLS
    ===================================================== */

const brush = document.getElementById("brush");

const eraser = document.getElementById("eraser");

const colorPicker = document.getElementById("color");

const sizePicker = document.getElementById("size");

const clear = document.getElementById("clear");

brush.addEventListener("click", () => {
  erasing = false;

  brush.classList.add("active");
  eraser.classList.remove("active");
});

eraser.addEventListener("click", () => {
  erasing = true;

  eraser.classList.add("active");
  brush.classList.remove("active");
});

colorPicker.addEventListener("input", (event) => {
  color = event.target.value;

  erasing = false;

  brush.classList.add("active");
  eraser.classList.remove("active");
});

sizePicker.addEventListener("input", (event) => {
  size = Number(event.target.value);
});

clear.addEventListener("click", () => {
  ctx.fillStyle = "#ffffff";

  ctx.fillRect(0, 0, canvas.width, canvas.height);
});

/* =====================================================
       CHAT
    ===================================================== */

const chatForm = document.getElementById("chatForm");

const chatInput = document.getElementById("chatInput");

const messages = document.getElementById("messages");

function escapeHTML(value) {
  const div = document.createElement("div");

  div.textContent = value;

  return div.innerHTML;
}

function addMessage(username, text, usernameColor = "#8b5cf6") {
  const message = document.createElement("div");

  message.className = "message";

  message.innerHTML = `
        <div
          class="message-name"
          style="color:${usernameColor}"
        >
          ${escapeHTML(username)}
        </div>

        <div class="bubble">
          ${escapeHTML(text)}
        </div>
      `;

  messages.appendChild(message);

  messages.scrollTop = messages.scrollHeight;
}

chatForm.addEventListener("submit", (event) => {
  event.preventDefault();

  const text = chatInput.value.trim();

  if (!text) return;

  addMessage("You", text, "#8b5cf6");

  chatInput.value = "";

  setTimeout(() => {
    const replies = [
      "Hmm... maybe! 👀",
      "Nope! Keep guessing 😆",
      "You're getting warmer! 🔥",
      "Interesting guess!",
      "I have no idea 😂",
    ];

    const reply = replies[Math.floor(Math.random() * replies.length)];

    addMessage("Sarah", reply, "#ec4899");
  }, 700);
});

/* =====================================================
       TIMER
    ===================================================== */

const timer = document.getElementById("timer");

let seconds = 72;

const timerInterval = setInterval(() => {
  seconds--;

  timer.innerHTML = `⏱ <span>${seconds}</span>s`;

  if (seconds <= 10) {
    timer.classList.add("danger");
  }

  if (seconds <= 0) {
    clearInterval(timerInterval);

    timer.innerHTML = "⏰ 0s";

    addMessage("Game", "Time's up! 🎉", "#f59e0b");
  }
}, 1000);

/* =====================================================
       MOCK PLAYER ANIMATION
    ===================================================== */

const statuses = [
  "🤔 Guessing",
  "💭 Thinking...",
  "🔥 Almost!",
  "👀 Watching",
  "😂 What is this?",
];

const players = document.querySelectorAll(".player");

setInterval(() => {
  const index = Math.floor(Math.random() * players.length);

  const status = players[index].querySelector(".player-status");

  if (status) {
    status.textContent = statuses[Math.floor(Math.random() * statuses.length)];
  }
}, 1700);

/* =====================================================
       INIT
    ===================================================== */

setTimeout(resizeCanvas, 100);

console.log("HERE");

// Wait for the browser to load the page content first
window.addEventListener("DOMContentLoaded", () => {
  const sessionId = "test-uuid-" + Math.floor(Math.random() * 1000);
  const socket = new WebSocket(`ws://localhost:8080/ws?session_id=${sessionId}`);

socket.onopen = () => {
    console.log(`Connected to Go Server! ✅ (Session: ${sessionId})`);

    // Send initial handshake event as JSON
    const initEvent = {
      type: "CHAT_MESSAGE",
      payload: {
        message: "Hello from client " + sessionId
      }
    };
    socket.send(JSON.stringify(initEvent));
  };

  socket.onmessage = (event) => {
    // writePump can batch multiple messages separated by \n
    const rawMessages = event.data.split("\n");

    rawMessages.forEach((rawMsg) => {
      if (!rawMsg.trim()) return;

      try {
        const parsed = JSON.parse(rawMsg);
        console.log("Parsed event from server:", parsed);
      } catch (err) {
        console.log("Raw message from server:", rawMsg);
      }
    });
  };

  socket.onerror = (err) => {
    console.error("WebSocket Error ❌:", err);
  };

  socket.onclose = (event) => {
    console.warn(`WebSocket Disconnected ⚠️ (Code: ${event.code}, Reason: ${event.reason})`);
  };

});
