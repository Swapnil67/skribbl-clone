/* =====================================================
       CANVAS
    ===================================================== */

// --- Global State ---
const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d");

let socket = null;
let isMyTurn = false;
let drawing = false;
let lastX = 0,
  lastY = 0;
let currentDrawerId = null;
let currentRoundNumber = null,
  maxRounds = null;

let color = "#25232b";
let size = 5;
let erasing = false;

const wordCard = document.querySelector(".word-card");
const drawingLabel = document.querySelector(".drawing-label");
const timer = document.getElementById("timer");

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
  updateCurrentRound();
});

function updateCurrentRound() {
  if (currentRoundNumber && maxRounds)
    document.querySelector(".round-pill").textContent =
      `🎮 Round ${currentRoundNumber} / ${maxRounds}`;
}

function initCanvasSync() {
  const roomId = getQueryParam("room_id") || "45S7HL";
  resizeCanvas();
  connectWebSocket(roomId);
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
  console.log(packet);

  // * Prevent echoing our own actions back onto the canvas
  if (sender_id == sessionId) return;

  switch (type) {
    case "PHASE_CHANGE":
      handlePhaseChange(payload);
      break;

    case "DRAWER_SECRET_WORD":
      // * The drawer receives and displays the actual word
      if (!isMyTurn) return;

      let wordEle = document.querySelector(".word");
      if (!wordEle) {
        wordEle = document.createElement("div");
        wordEle.classList.add("word");
        wordCard.appendChild(wordEle);
      }
      wordEle.innerText = `${payload.word.toUpperCase()}`;
      break;

    case "WORD_SELECTED":
      // * Hide this from the drawer
      if (isMyTurn) return;
      const { word_length } = payload;
      const guessWordDiv = document.createElement("div");
      guessWordDiv.classList.add("word");
      for (let i = 0; i < word_length; i++) {
        const span = document.createElement("span");
        span.classList.add("hidden");
        span.textContent = "_";
        guessWordDiv.appendChild(span);
      }
      wordCard.appendChild(guessWordDiv);
      break;

    case "TIMER_TICK":
      timer.innerHTML = `⏱ <span>${payload?.remaining_seconds}</span>s`;
      break;

    case "DRAW_STROKE":
      if (sender_id != sessionId) {
        renderStroke(payload);
      }
      break;

    case "CLEAR_CANVAS":
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      break;

    // case "UNDO_STROKE":
    //   console.log("Undo triggered by:", sender_id);
    //   break;

    // * --- Guess & Chat Events ---
    case "CHAT_MESSAGE":
      addMessage(payload.username, payload.text);
      break;

    case "PLAYER_GUESSED":
      // appendSystemMessage(`🎉 ${payload.username} guessed the word! (+${payload.points_earned} pts)`, "system-success");
      // markPlayerGuessed(payload.session_id);
      break;

    case "CLOSE_GUESS_ALERT":
      // appendSystemMessage(`⚠️ ${payload.message}`, "system-close");
      break;

    case "SYSTEM_ALERT":
      // appendSystemMessage(`🔔 ${payload.message}`, "system-warning");
      break;

    case "SCORE_UPDATE":
      // renderScoreboard(payload.scores);
      break;
  }
}

function handlePhaseChange(payload) {
  const { phase, current_drawer_id } = payload;

  isMyTurn = payload.current_drawer_id === sessionId;

  console.log("phase ", phase);

  if (phase == "WORD_SELECTION") {
    document.querySelector(".word-label").textContent = "";
    document.querySelector(".word")?.remove();

    const phaseTitleEle = document.createElement("div");
    phaseTitleEle.id = "phaseTitle";
    phaseTitleEle.classList.add("system-message");
    let message = isMyTurn
      ? "Your turn to draw! Choosing word..."
      : `${current_drawer_id} is selecting the word`;
    phaseTitleEle.textContent = message;
    wordCard.insertAdjacentElement("afterbegin", phaseTitleEle);

    currentDrawerId = current_drawer_id;
    currentRoundNumber = payload?.round_number;
    maxRounds = payload?.max_rounds;
    updateCurrentRound();
  } else if (phase == "DRAWING") {
    const phaseTitleEle = document.querySelector("#phaseTitle");
    phaseTitleEle?.remove();
    const wordLabel = document.querySelector(".word-label");

    if (isMyTurn) {
      // 🎨 Drawer State
      canvas.style.cursor = "crosshair";
      wordLabel.textContent = "You are drawing...";
    } else {
      // 🔍 Guesser State
      canvas.style.cursor = "not-allowed";
      wordLabel.textContent = `${current_drawer_id} is drawing...`;
    }

    drawingLabel.textContent = `✏️ ${current_drawer_id} canvas`;
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
  erasing = true;
}

function setBrush() {
  erasing = false;
}

function clearCanvas() {
  if (!isMyTurn) return;
  ctx.fillStyle = "#ffffff";
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  sendEvent("CLEAR_CANVAS", {});
}

function setLineWidth(e) {
  size = Number(e.target.value);
}

const brush = document.getElementById("brush");
const eraser = document.getElementById("eraser");
const colorPicker = document.getElementById("color");
const sizePicker = document.getElementById("size");
const clear = document.getElementById("clear");
clear.addEventListener("click", clearCanvas);
sizePicker.addEventListener("input", setLineWidth);

brush.addEventListener("click", () => {
  setBrush();
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
    username: sessionId,
    text,
  });
});
