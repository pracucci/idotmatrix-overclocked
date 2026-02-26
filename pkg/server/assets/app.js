// iDotMatrix Console - App Logic

document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initButtons();
    initFileUploads();
    initPreview();
    initSnake();
    startStatusPolling();
});

// Tab Switching
function initTabs() {
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => {
            // Update tab buttons
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            // Update tab content
            const tabId = tab.dataset.tab;
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.toggle('active', content.id === `tab-${tabId}`);
            });
        });
    });
}

// Button Actions
function initButtons() {
    // Power buttons
    document.getElementById('btn-power-on').addEventListener('click', () => {
        apiPost('/api/power/on');
    });

    document.getElementById('btn-power-off').addEventListener('click', () => {
        apiPost('/api/power/off');
        clearPreview();
    });

    // Action buttons
    document.querySelectorAll('[data-action]').forEach(btn => {
        btn.addEventListener('click', () => {
            const action = btn.dataset.action;
            handleAction(action, btn);
        });
    });

    // Enter key submits the card's action
    document.querySelectorAll('.card').forEach(card => {
        card.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                const btn = card.querySelector('[data-action]');
                if (btn && !btn.disabled) {
                    e.preventDefault();
                    btn.click();
                }
            }
        });
    });
}

async function handleAction(action, btn) {
    setButtonLoading(btn, true);

    // Update preview when action is triggered
    updatePreviewForAction(action);

    try {
        let result;

        switch (action) {
            case 'text':
                result = await apiPost('/api/text', {
                    text: document.getElementById('text-input').value,
                    animation: document.getElementById('text-animation').value,
                    color: document.getElementById('text-color').value
                });
                break;

            case 'emoji':
                result = await apiPost('/api/emoji', {
                    name: document.getElementById('emoji-name').value
                });
                break;

            case 'grot':
                result = await apiPost('/api/grot', {
                    name: document.getElementById('grot-name').value
                });
                break;

            case 'fire':
                result = await apiPost('/api/fire');
                break;

            case 'clock':
                result = await apiPost('/api/clock', {
                    style: document.getElementById('clock-style').value,
                    show_date: document.getElementById('clock-show-date').checked,
                    hour_24: document.getElementById('clock-24h').checked,
                    color: document.getElementById('clock-color').value
                });
                break;

            case 'showimage':
                result = await uploadFile('/api/showimage', document.getElementById('image-file').files[0]);
                break;

            case 'showgif':
                result = await uploadFile('/api/showgif', document.getElementById('gif-file').files[0]);
                break;

            case 'snake':
                // Snake is handled separately by initSnake()
                return;

            default:
                throw new Error(`Unknown action: ${action}`);
        }

        if (result.success) {
            showNotification('Success!', 'success');
        } else {
            showNotification(result.error || 'Unknown error', 'error');
        }
    } catch (err) {
        showNotification(err.message, 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function updatePreviewForAction(action) {
    switch (action) {
        case 'text':
            const text = document.getElementById('text-input').value;
            if (text) {
                updatePreview('/api/text.gif', {
                    text: text,
                    animation: document.getElementById('text-animation').value,
                    color: document.getElementById('text-color').value
                });
            } else {
                clearPreview();
            }
            break;

        case 'emoji':
            updatePreview('/api/emoji.gif', {
                name: document.getElementById('emoji-name').value
            });
            break;

        case 'grot':
            updatePreview('/api/grot.gif', {
                name: document.getElementById('grot-name').value
            });
            break;

        case 'fire':
            updatePreview('/api/fire.gif', {});
            break;

        default:
            clearPreview();
            break;
    }
}

// API Helpers
const IMAGE_ENDPOINTS = ['/api/text', '/api/emoji', '/api/grot', '/api/fire', '/api/showimage', '/api/showgif'];

async function apiPost(endpoint, params = {}) {
    const url = new URL(endpoint, window.location.origin);

    // Add mirrored and brightness for image endpoints
    if (IMAGE_ENDPOINTS.some(e => endpoint.startsWith(e))) {
        if (document.getElementById('mirrored-checkbox').checked) {
            params.mirrored = 'true';
        }
        params.brightness = document.getElementById('brightness-slider').value;
    }

    Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
            url.searchParams.set(key, value);
        }
    });

    const response = await fetch(url, { method: 'POST' });
    return response.json();
}

async function uploadFile(endpoint, file) {
    if (!file) {
        throw new Error('No file selected');
    }

    const formData = new FormData();
    formData.append('file', file);

    const url = new URL(endpoint, window.location.origin);
    if (document.getElementById('mirrored-checkbox').checked) {
        url.searchParams.set('mirrored', 'true');
    }
    url.searchParams.set('brightness', document.getElementById('brightness-slider').value);

    const response = await fetch(url, {
        method: 'POST',
        body: formData
    });
    return response.json();
}

// Status Display
let wasConnected = true;

function checkStatus() {
    const statusEl = document.getElementById('status');

    fetch('/api/status')
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                statusEl.textContent = 'Connected';
                statusEl.className = 'status connected';
                if (!wasConnected) {
                    showNotification('Device reconnected', 'success');
                }
                wasConnected = true;
            } else {
                statusEl.textContent = 'Reconnecting...';
                statusEl.className = 'status disconnected';
                if (wasConnected) {
                    showNotification('Device disconnected, reconnecting...', 'error');
                }
                wasConnected = false;
            }
        })
        .catch(() => {
            statusEl.textContent = 'Disconnected';
            statusEl.className = 'status disconnected';
            wasConnected = false;
        });
}

function startStatusPolling() {
    checkStatus();
    setInterval(checkStatus, 3000);
}

function setButtonLoading(btn, loading) {
    btn.disabled = loading;
}

function showNotification(message, type = 'success') {
    const container = document.getElementById('notifications');

    // Clear previous error notifications
    container.querySelectorAll('.notification.error').forEach(n => n.remove());

    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;

    // Click to dismiss
    notification.addEventListener('click', () => notification.remove());

    container.appendChild(notification);

    // Auto-dismiss success after 3s
    if (type === 'success') {
        setTimeout(() => notification.remove(), 3000);
    }
}

// File Uploads
function initFileUploads() {
    setupDropZone('image-drop-zone', 'image-file', 'image-preview', 'image-preview-container', 'image-filename', 'btn-showimage');
    setupDropZone('gif-drop-zone', 'gif-file', 'gif-preview', 'gif-preview-container', 'gif-filename', 'btn-showgif');
}

function setupDropZone(zoneId, inputId, previewId, containerId, filenameId, btnId) {
    const zone = document.getElementById(zoneId);
    const input = document.getElementById(inputId);
    const preview = document.getElementById(previewId);
    const container = document.getElementById(containerId);
    const filename = document.getElementById(filenameId);
    const btn = document.getElementById(btnId);

    // Click to browse
    zone.addEventListener('click', () => input.click());

    // Drag events
    zone.addEventListener('dragover', (e) => {
        e.preventDefault();
        zone.classList.add('dragover');
    });

    zone.addEventListener('dragleave', () => {
        zone.classList.remove('dragover');
    });

    zone.addEventListener('drop', (e) => {
        e.preventDefault();
        zone.classList.remove('dragover');

        const files = e.dataTransfer.files;
        if (files.length > 0) {
            input.files = files;
            handleFileSelect(input, preview, container, filename, btn);
        }
    });

    // File input change
    input.addEventListener('change', () => {
        handleFileSelect(input, preview, container, filename, btn);
    });
}

function handleFileSelect(input, preview, container, filenameEl, btn) {
    const file = input.files[0];
    if (!file) {
        container.hidden = true;
        btn.disabled = true;
        return;
    }

    // Show preview
    const reader = new FileReader();
    reader.onload = (e) => {
        preview.src = e.target.result;
        filenameEl.textContent = file.name;
        container.hidden = false;
        btn.disabled = false;
    };
    reader.readAsDataURL(file);
}

// Server-side GIF Preview
const DEBOUNCE_MS = 300;
let previewDebounceTimer = null;

function initPreview() {
    // Restore settings from localStorage
    const mirroredCheckbox = document.getElementById('mirrored-checkbox');
    const brightnessSlider = document.getElementById('brightness-slider');
    const brightnessValue = document.getElementById('brightness-value');

    const savedMirrored = localStorage.getItem('idotmatrix-mirrored');
    if (savedMirrored !== null) {
        mirroredCheckbox.checked = savedMirrored === 'true';
    }

    const savedBrightness = localStorage.getItem('idotmatrix-brightness');
    if (savedBrightness !== null) {
        brightnessSlider.value = savedBrightness;
        brightnessValue.textContent = savedBrightness;
    }

    // Start with empty preview
    clearPreview();

    // Text preview with debounce
    const textInput = document.getElementById('text-input');
    const textAnimation = document.getElementById('text-animation');
    const textColor = document.getElementById('text-color');

    function updateTextPreview() {
        const text = textInput.value;
        if (!text) {
            clearPreview();
            return;
        }
        updatePreview('/api/text.gif', {
            text: text,
            animation: textAnimation.value,
            color: textColor.value
        });
    }

    textInput.addEventListener('input', () => {
        clearTimeout(previewDebounceTimer);
        previewDebounceTimer = setTimeout(updateTextPreview, DEBOUNCE_MS);
    });
    textAnimation.addEventListener('change', updateTextPreview);
    textColor.addEventListener('change', updateTextPreview);

    // Emoji preview
    const emojiSelect = document.getElementById('emoji-name');
    emojiSelect.addEventListener('change', () => {
        updatePreview('/api/emoji.gif', { name: emojiSelect.value });
    });

    // Grot preview
    const grotSelect = document.getElementById('grot-name');
    grotSelect.addEventListener('change', () => {
        updatePreview('/api/grot.gif', { name: grotSelect.value });
    });

    // Mirror checkbox - refresh preview when toggled and save to localStorage
    mirroredCheckbox.addEventListener('change', () => {
        localStorage.setItem('idotmatrix-mirrored', mirroredCheckbox.checked);
        refreshCurrentPreview();
    });

    // Brightness slider - refresh preview when changed and save to localStorage
    brightnessSlider.addEventListener('input', () => {
        brightnessValue.textContent = brightnessSlider.value;
        localStorage.setItem('idotmatrix-brightness', brightnessSlider.value);
        refreshCurrentPreview();
    });
}

// Store the last preview URL (without mirrored/brightness/timestamp)
let lastPreviewUrl = null;

// Refresh the current preview by updating mirrored/brightness on the last URL
function refreshCurrentPreview() {
    if (!lastPreviewUrl) return;

    const previewImg = document.getElementById('preview');
    const url = new URL(lastPreviewUrl);

    // Update mirrored parameter
    if (document.getElementById('mirrored-checkbox').checked) {
        url.searchParams.set('mirrored', 'true');
    } else {
        url.searchParams.delete('mirrored');
    }

    // Update brightness parameter
    url.searchParams.set('brightness', document.getElementById('brightness-slider').value);

    // Add timestamp to prevent caching
    url.searchParams.set('_t', Date.now());

    previewImg.src = url.toString();
}

function updatePreview(endpoint, params) {
    const previewImg = document.getElementById('preview');
    const url = new URL(endpoint, window.location.origin);

    // Add content params (not mirrored/brightness)
    Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
            url.searchParams.set(key, value);
        }
    });

    // Store the base URL (without mirrored/brightness/timestamp)
    lastPreviewUrl = url.toString();

    // Add mirrored parameter
    if (document.getElementById('mirrored-checkbox').checked) {
        url.searchParams.set('mirrored', 'true');
    }

    // Add brightness parameter
    url.searchParams.set('brightness', document.getElementById('brightness-slider').value);

    // Add timestamp to prevent caching
    url.searchParams.set('_t', Date.now());

    previewImg.src = url.toString();
}

function clearPreview() {
    const previewImg = document.getElementById('preview');
    // Set to a 1x1 transparent GIF
    previewImg.src = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';
    lastPreviewUrl = null;
}

// Snake Game
let snakeActive = false;
let snakeSessionId = null;
let snakeStatusPollInterval = null;

function initSnake() {
    const startBtn = document.getElementById('btn-snake-start');
    const stopBtn = document.getElementById('btn-snake-stop');

    startBtn.addEventListener('click', async () => {
        startBtn.disabled = true;

        try {
            const result = await apiPost('/api/snake/start');
            if (result.success) {
                snakeActive = true;
                snakeSessionId = result.session_id;
                updateSnakeUI(true);
                showNotification('Game started! Use arrow keys or WASD to move', 'success');
                startSnakeStatusPoll();
            } else {
                showNotification(result.error || 'Failed to start game', 'error');
                startBtn.disabled = false;
            }
        } catch (err) {
            showNotification(err.message, 'error');
            startBtn.disabled = false;
        }
    });

    stopBtn.addEventListener('click', async () => {
        stopBtn.disabled = true;

        try {
            await apiPost('/api/snake/stop');
            snakeActive = false;
            snakeSessionId = null;
            updateSnakeUI(false);
            showNotification('Game stopped', 'success');
            stopSnakeStatusPoll();
        } catch (err) {
            showNotification(err.message, 'error');
            stopBtn.disabled = false;
        }
    });

    // Global keyboard listener for snake controls
    document.addEventListener('keydown', handleSnakeKeydown);

    // Check if there's already an active game
    checkSnakeStatus();
}

function updateSnakeUI(active) {
    const startBtn = document.getElementById('btn-snake-start');
    const stopBtn = document.getElementById('btn-snake-stop');
    const hintEl = document.getElementById('snake-hint');
    const card = document.getElementById('snake-card');

    startBtn.disabled = active;
    stopBtn.disabled = !active;

    if (active) {
        hintEl.textContent = 'Use arrow keys or WASD to move. Q to quit.';
        card.classList.add('snake-active');
    } else {
        hintEl.textContent = 'Click Start Game to begin';
        card.classList.remove('snake-active');
    }
}

async function handleSnakeKeydown(e) {
    if (!snakeActive) return;

    // Map browser keys to API keys
    const keyMap = {
        'ArrowUp': 'up',
        'ArrowDown': 'down',
        'ArrowLeft': 'left',
        'ArrowRight': 'right',
        'w': 'up',
        'W': 'up',
        's': 'down',
        'S': 'down',
        'a': 'left',
        'A': 'left',
        'd': 'right',
        'D': 'right',
        'q': 'quit',
        'Q': 'quit',
        'r': 'restart',
        'R': 'restart',
        'Enter': 'restart',
        ' ': 'restart'
    };

    const apiKey = keyMap[e.key];
    if (!apiKey) return;

    // Prevent default for arrow keys (scrolling) and space (scrolling)
    if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', ' '].includes(e.key)) {
        e.preventDefault();
    }

    try {
        const result = await apiPost('/api/snake/input', { key: apiKey });
        if (result.success) {
            updateSnakeState(result.state);
        } else if (result.error === 'no active game session') {
            snakeActive = false;
            updateSnakeUI(false);
            stopSnakeStatusPoll();
        }
    } catch (err) {
        console.error('Failed to send snake input:', err);
    }
}

function updateSnakeState(state) {
    switch (state) {
        case 'cover':
            showNotification('Press any key to start playing', 'success');
            break;
        case 'playing':
            showNotification('Playing - use arrow keys or WASD', 'success');
            break;
        case 'gameover':
            showNotification('Game Over! Press R or Enter to restart', 'error');
            break;
        case 'ended':
            snakeActive = false;
            updateSnakeUI(false);
            showNotification('Game ended', 'success');
            stopSnakeStatusPoll();
            break;
    }
}

function startSnakeStatusPoll() {
    stopSnakeStatusPoll();
    snakeStatusPollInterval = setInterval(checkSnakeStatus, 2000);
}

function stopSnakeStatusPoll() {
    if (snakeStatusPollInterval) {
        clearInterval(snakeStatusPollInterval);
        snakeStatusPollInterval = null;
    }
}

async function checkSnakeStatus() {
    try {
        const response = await fetch('/api/snake/status');
        const result = await response.json();

        if (result.success && result.state !== 'ended') {
            if (!snakeActive) {
                snakeActive = true;
                updateSnakeUI(true);
                startSnakeStatusPoll();
            }
            updateSnakeState(result.state);
        } else if (snakeActive) {
            snakeActive = false;
            updateSnakeUI(false);
            stopSnakeStatusPoll();
        }
    } catch (err) {
        // Ignore errors during status check
    }
}
