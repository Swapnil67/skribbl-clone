/* =====================================================
       CANVAS
    ===================================================== */

    // --- Global State ---
const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d");

let socket = null;
let drawing = false;
let lastX = 0, lastY = 0;

let color = "#25232b";
let size = 5;
let erasing = false;

// Unique session identifier for this tab
const sessionId = "user-" + Math.floor(Math.random() * 10000);

// * Extract URL query params (e.g., ?room_id=ABC123)
function getQueryParam(param) {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get(param);
}

/* =====================================================
       INIT
    ===================================================== */

// Initialize on page load
window.addEventListener("DOMContentLoaded", () => {
  initCanvasSync("paintCanvas");
});

function initCanvasSync() {
  const roomId = getQueryParam("room_id") || "DEFAULT";

  resizeCanvas()
  connectWebSocket(roomId)
}

function connectWebSocket(roomId) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  const wsUrl = `${protocol}//${host}/ws?room_id=${encodeURIComponent(roomId)}&session_id=${encodeURIComponent(sessionId)}`;
  socket = new WebSocket(wsUrl);

  socket.onopen = () => {
    console.log(`Connected to Room: [${roomId}] as Session: [${sessionId}] ✅`);
  };

  socket.onmessage = (event) => {
    // writePump can batch multiple messages separated by \n
    const rawMessages = event.data.split("\n");

    rawMessages.forEach((rawMsg) => {
      if (!rawMsg.trim()) return;

      try {
        const packet = JSON.parse(rawMsg);
        console.log("Parsed event from server:", packet);
        handleIncomingEvent(packet);
      } catch (err) {
        console.log("Raw message from server:", err);
      }
    });
  };

  socket.onerror = (err) => {
    console.error("WebSocket Error ❌:", err);
  };

  socket.onclose = (event) => {
    console.warn(
      `WebSocket Disconnected ⚠️ (Code: ${event.code}, Reason: ${event.reason})`,
    );
  };
}

// --- 2. Outbound Event Dispatcher ---
function sendEvent(type, payload) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;
  socket.send(JSON.stringify({ type, payload }));
}

// --- 3. Inbound Event Router ---
function handleIncomingEvent(packet) {
  const { type, sender_id, payload } = packet;

  // * Prevent echoing our own actions back onto the canvas
  if (sender_id == sessionId) return;

  switch (type) {
    case "DRAW_STROKE":
      renderStroke(payload);
      break;

    case "CLEAR_CANVAS":
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      break;

    // case "UNDO_STROKE":
    //   console.log("Undo triggered by:", sender_id);
    //   break;

    case "CHAT_MESSAGE":
      addMessage(payload.username, payload.text)
      break;
  }
}

function renderStroke(stroke) {
  ctx.save();
  ctx.beginPath();
  ctx.moveTo(stroke.prev_x, stroke.prev_y);
  ctx.lineTo(stroke.curr_x, stroke.curr_y);

  if (stroke.mode == "eraser") {
    ctx.globalCompositeOperation = "destination-out";
    ctx.lineWidth = stroke.line_width;
  } else {
    ctx.globalCompositeOperation = "source-over";
    ctx.strokeStyle = stroke?.currentColor;
    ctx.lineWidth = stroke?.line_width;
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
  }

  ctx.stroke();
  ctx.restore();
}

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

  // * Draw Locally
  let currentColor = erasing ? "#ffffff" : color;
  let mode = erasing ? "eraser" : "pencil";
  const strokeData = {
    prev_x: lastX,
    prev_y: lastY,
    curr_x: p.x,
    curr_y: p.y,
    color: currentColor,
    line_width: size,
    mode,
  };
  renderStroke(strokeData);

  // * Broadcast stroke to everyone else
  sendEvent("DRAW_STROKE", strokeData);

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

function setEraser() {
  erasing = true
}

function setBrush() {
  erasing = false
}

function clearCanvas() {
  ctx.fillStyle = "#ffffff";
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  sendEvent("CLEAR_CANVAS", {});
}

function setLineWidth(e) {
  size = Number(e.target.value);;
}

const brush = document.getElementById("brush");
const eraser = document.getElementById("eraser");
const colorPicker = document.getElementById("color");
const sizePicker = document.getElementById("size");
const clear = document.getElementById("clear");
clear.addEventListener('click', clearCanvas)
sizePicker.addEventListener('input', setLineWidth)

brush.addEventListener("click", () => {
  setBrush()
  brush.classList.add("active");
  eraser.classList.remove("active");
});

eraser.addEventListener("click", () => {
  setEraser();
  eraser.classList.add("active");
  brush.classList.remove("active");
});

colorPicker.addEventListener("input", (event) => {
  color = event.target.value;
  erasing = false;
  brush.classList.add("active");
  eraser.classList.remove("active");
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
        <div class="message-name"style="color:${usernameColor}">
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
  if (!text || !text.length) return;
  addMessage("You", text, "#8b5cf6");
  chatInput.value = "";

  sendEvent("CHAT_MESSAGE", {
    text,
    IsGuess: false,
    username: sessionId,
  });
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