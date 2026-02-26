package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/emoji"
	"github.com/pracucci/idotmatrix-overclocked/pkg/fire"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/grot"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/pracucci/idotmatrix-overclocked/pkg/text"
)

// apiResponse is the standard API response format.
type apiResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// snakeStartResponse is the response for starting a snake game.
type snakeStartResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// snakeInputResponse is the response for sending snake input.
type snakeInputResponse struct {
	Success bool   `json:"success"`
	State   string `json:"state,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) registerAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/power/on", s.handlePowerOn)
	mux.HandleFunc("/api/power/off", s.handlePowerOff)
	mux.HandleFunc("/api/text", s.handleText)
	mux.HandleFunc("/api/emoji", s.handleEmoji)
	mux.HandleFunc("/api/grot", s.handleGrot)
	mux.HandleFunc("/api/fire", s.handleFire)
	mux.HandleFunc("/api/clock", s.handleClock)
	mux.HandleFunc("/api/showimage", s.handleShowImage)
	mux.HandleFunc("/api/showgif", s.handleShowGIF)
	mux.HandleFunc("/api/snake/start", s.handleSnakeStart)
	mux.HandleFunc("/api/snake/input", s.handleSnakeInput)
	mux.HandleFunc("/api/snake/stop", s.handleSnakeStop)
	mux.HandleFunc("/api/snake/status", s.handleSnakeStatus)

	// Preview endpoints (return GIF without sending to device)
	mux.HandleFunc("/api/text.gif", s.handleTextPreview)
	mux.HandleFunc("/api/emoji.gif", s.handleEmojiPreview)
	mux.HandleFunc("/api/grot.gif", s.handleGrotPreview)
	mux.HandleFunc("/api/fire.gif", s.handleFirePreview)
}

func writeJSON(w http.ResponseWriter, resp apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// parseBrightness extracts the brightness parameter from the request.
// Returns 100 (no change) if not specified or invalid, otherwise 0-100.
func parseBrightness(r *http.Request) int {
	if b := r.URL.Query().Get("brightness"); b != "" {
		if val, err := strconv.Atoi(b); err == nil && val >= 0 && val <= 100 {
			return val
		}
	}
	return 100
}

// parseMirrored extracts the mirrored parameter from the request.
func parseMirrored(r *http.Request) bool {
	return r.URL.Query().Get("mirrored") == "true"
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if s.IsConnected() {
		writeJSON(w, apiResponse{Success: true})
	} else {
		writeJSON(w, apiResponse{Success: false, Error: "device disconnected, reconnecting..."})
	}
}

func (s *Server) handlePowerOn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	err := s.withDevice(func(d *protocol.Device) error {
		if err := protocol.SetPowerState(d, true); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Power on failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handlePowerOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	err := s.withDevice(func(d *protocol.Device) error {
		if err := protocol.SetPowerState(d, false); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Power off failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	textMsg := r.URL.Query().Get("text")
	if textMsg == "" {
		writeJSON(w, apiResponse{Success: false, Error: "text parameter required"})
		return
	}

	animation := r.URL.Query().Get("animation")
	if animation == "" {
		animation = "none"
	}

	colorName := r.URL.Query().Get("color")
	if colorName == "" {
		colorName = "white"
	}

	// Validate color
	colorName = strings.ToLower(strings.TrimSpace(colorName))
	color, ok := graphic.ColorPalette[colorName]
	if !ok {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("unknown color: %s (valid: %s)", colorName, strings.Join(graphic.ColorNames(), ", "))})
		return
	}

	// Convert text to uppercase
	msg := strings.ToUpper(textMsg)

	// Validate text fits
	lines := text.WrapText(msg)
	blockHeight := text.TextBlockHeight(lines)
	if blockHeight > graphic.DisplayHeight {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("text too long: wrapped to %d lines (%d pixels, max %d)", len(lines), blockHeight, graphic.DisplayHeight)})
		return
	}

	// Generate image
	opts := text.DefaultAnimationOptions()
	opts.TextOptions.TextColor = color
	opts.TextOptions.ShadowColor = graphic.ShadowFor(color)

	img, errMsg := text.GenerateAnimation(animation, msg, opts)
	if errMsg != "" {
		writeJSON(w, apiResponse{Success: false, Error: errMsg})
		return
	}

	if parseMirrored(r) {
		img = img.Mirror()
	}

	img = img.AdjustBrightness(parseBrightness(r))

	err := s.withDevice(func(d *protocol.Device) error {
		if img.Type == graphic.ImageTypeStatic {
			if err := protocol.SetDrawMode(d, 1); err != nil {
				return err
			}
			rawBytes, err := img.RawBytes()
			if err != nil {
				return err
			}
			if err := protocol.SendImage(d, rawBytes); err != nil {
				return err
			}
		} else {
			gifBytes, err := img.GIFBytes()
			if err != nil {
				return err
			}
			if err := protocol.SendGIF(d, gifBytes, s.logger); err != nil {
				return err
			}
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Text display failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleEmoji(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("name parameter required (available: %s)", strings.Join(emoji.Names(), ", "))})
		return
	}

	img, err := emoji.Generate(name)
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	if parseMirrored(r) {
		img = img.Mirror()
	}

	img = img.AdjustBrightness(parseBrightness(r))

	err = s.withDevice(func(d *protocol.Device) error {
		gifBytes, err := img.GIFBytes()
		if err != nil {
			return err
		}
		if err := protocol.SendGIF(d, gifBytes, s.logger); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Emoji display failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleGrot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("name parameter required (available: %s)", strings.Join(grot.Names(), ", "))})
		return
	}

	img, err := grot.Generate(name)
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	if parseMirrored(r) {
		img = img.Mirror()
	}

	img = img.AdjustBrightness(parseBrightness(r))

	err = s.withDevice(func(d *protocol.Device) error {
		gifBytes, err := img.GIFBytes()
		if err != nil {
			return err
		}
		level.Info(s.logger).Log("msg", "Uploading GIF to device", "name", name)
		if err := protocol.SendGIF(d, gifBytes, s.logger); err != nil {
			return err
		}
		level.Info(s.logger).Log("msg", "GIF upload complete", "name", name)
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Grot display failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleFire(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	gifData := fire.GenerateGIF()

	g, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to decode GIF: " + err.Error()})
		return
	}
	if parseMirrored(r) {
		g = graphic.MirrorGIFHorizontal(g)
	}
	g = graphic.AdjustBrightnessGIF(g, parseBrightness(r))
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to re-encode GIF: " + err.Error()})
		return
	}
	gifData = buf.Bytes()

	err = s.withDevice(func(d *protocol.Device) error {
		if err := protocol.SendGIF(d, gifData, s.logger); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Fire display failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleClock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	// Parse style (default: 4 = HourGlass)
	style := protocol.ClockAnimatedHourGlass
	if styleStr := r.URL.Query().Get("style"); styleStr != "" {
		var err error
		style, err = strconv.Atoi(styleStr)
		if err != nil || style < 0 || style > protocol.ClockAnimatedHourGlass {
			writeJSON(w, apiResponse{Success: false, Error: "invalid style (0-4)"})
			return
		}
	}

	// Parse show_date (default: true)
	showDate := true
	if showDateStr := r.URL.Query().Get("show_date"); showDateStr != "" {
		showDate = showDateStr == "true" || showDateStr == "1"
	}

	// Parse hour_24 (default: true)
	hour24 := true
	if hour24Str := r.URL.Query().Get("hour_24"); hour24Str != "" {
		hour24 = hour24Str == "true" || hour24Str == "1"
	}

	// Parse color (default: white)
	colorName := r.URL.Query().Get("color")
	if colorName == "" {
		colorName = "white"
	}
	colorName = strings.ToLower(strings.TrimSpace(colorName))
	color, ok := graphic.ColorPalette[colorName]
	if !ok {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("unknown color: %s (valid: %s)", colorName, strings.Join(graphic.ColorNames(), ", "))})
		return
	}

	t := time.Now()

	err := s.withDevice(func(d *protocol.Device) error {
		if err := protocol.SetTime(d, t.Year(), int(t.Month()), t.Day(), int(t.Weekday())+1, t.Hour(), t.Minute(), t.Second()); err != nil {
			return err
		}
		if err := protocol.SetClockMode(d, style, showDate, hour24, graphic.Color{color[0], color[1], color[2]}); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Clock set failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

const (
	showgifDisplaySize    = 64
	showgifMaxFrames      = 64
	showgifMinFrameTimeMs = 16
)

func (s *Server) handleShowImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to parse form: " + err.Error()})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "file required"})
		return
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to decode image: " + err.Error()})
		return
	}

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	if width != 64 || height != 64 {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("image is %dx%d, expected 64x64", width, height)})
		return
	}

	// Convert to raw RGB
	rgbData := make([]byte, 64*64*3)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			red, green, blue, _ := img.At(x, y).RGBA()
			rgbData[idx] = uint8(red >> 8)
			rgbData[idx+1] = uint8(green >> 8)
			rgbData[idx+2] = uint8(blue >> 8)
			idx += 3
		}
	}

	if parseMirrored(r) {
		rgbData = graphic.MirrorBufferHorizontal(rgbData)
	}

	rgbData = graphic.AdjustBrightnessBuffer(rgbData, parseBrightness(r))

	err = s.withDevice(func(d *protocol.Device) error {
		if err := protocol.SetDrawMode(d, 1); err != nil {
			return err
		}
		if err := protocol.SendImage(d, rgbData); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "Image display failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleShowGIF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to parse form: " + err.Error()})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "file required"})
		return
	}
	defer file.Close()

	// Read file into buffer
	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to read file: " + err.Error()})
		return
	}

	// Decode GIF
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to decode GIF: " + err.Error()})
		return
	}

	// Validate dimensions
	if g.Config.Width != showgifDisplaySize || g.Config.Height != showgifDisplaySize {
		writeJSON(w, apiResponse{Success: false, Error: fmt.Sprintf("GIF is %dx%d, expected 64x64", g.Config.Width, g.Config.Height)})
		return
	}

	numFrames := len(g.Image)
	if numFrames == 0 {
		writeJSON(w, apiResponse{Success: false, Error: "GIF has no frames"})
		return
	}
	if numFrames > showgifMaxFrames {
		numFrames = showgifMaxFrames
	}

	// Calculate and adjust frame delays
	delays := make([]int, numFrames)
	for i := 0; i < numFrames; i++ {
		delay := g.Delay[i]
		if delay < showgifMinFrameTimeMs/10 {
			delay = showgifMinFrameTimeMs / 10
		}
		delays[i] = delay
	}

	// Re-composite and re-encode frames
	canvas := image.NewRGBA(image.Rect(0, 0, showgifDisplaySize, showgifDisplaySize))
	newFrames := make([]*image.Paletted, numFrames)

	for i := 0; i < numFrames; i++ {
		frame := g.Image[i]
		bounds := frame.Bounds()
		draw.Draw(canvas, bounds, frame, bounds.Min, draw.Over)
		palettedFrame := image.NewPaletted(image.Rect(0, 0, showgifDisplaySize, showgifDisplaySize), palette.Plan9)
		draw.Draw(palettedFrame, palettedFrame.Bounds(), canvas, image.Point{}, draw.Src)
		newFrames[i] = palettedFrame
		if i < len(g.Disposal) && g.Disposal[i] == gif.DisposalBackground {
			draw.Draw(canvas, bounds, image.Black, image.Point{}, draw.Src)
		}
	}

	// Re-encode GIF
	newGIF := &gif.GIF{
		Image:     newFrames,
		Delay:     delays,
		LoopCount: 0,
		Disposal:  make([]byte, numFrames),
	}
	for i := range newGIF.Disposal {
		newGIF.Disposal[i] = gif.DisposalBackground
	}

	if parseMirrored(r) {
		newGIF = graphic.MirrorGIFHorizontal(newGIF)
	}

	newGIF = graphic.AdjustBrightnessGIF(newGIF, parseBrightness(r))

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, newGIF); err != nil {
		writeJSON(w, apiResponse{Success: false, Error: "failed to re-encode GIF: " + err.Error()})
		return
	}

	gifData := buf.Bytes()

	err = s.withDevice(func(d *protocol.Device) error {
		if err := protocol.SendGIF(d, gifData, s.logger); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "GIF display failed", "err", err)
		writeJSON(w, apiResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleSnakeStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	// Check if a game is already running
	if s.snakeManager.HasActiveSession() {
		sessionID := s.snakeManager.GetSessionID()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snakeStartResponse{
			Success:   true,
			SessionID: sessionID,
		})
		return
	}

	// Start the game with direct device access (game manages its own rendering)
	sessionID, err := s.snakeManager.StartSession(s.device)
	if err != nil {
		level.Error(s.logger).Log("msg", "Failed to start snake game", "err", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snakeStartResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	level.Info(s.logger).Log("msg", "Snake game started", "session_id", sessionID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snakeStartResponse{
		Success:   true,
		SessionID: sessionID,
	})
}

func (s *Server) handleSnakeInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snakeInputResponse{
			Success: false,
			Error:   "key parameter required (up, down, left, right, quit, restart)",
		})
		return
	}

	// Validate key
	validKeys := map[string]bool{
		"up": true, "down": true, "left": true, "right": true,
		"quit": true, "restart": true,
	}
	if !validKeys[key] {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snakeInputResponse{
			Success: false,
			Error:   "invalid key (valid: up, down, left, right, quit, restart)",
		})
		return
	}

	state, ok := s.snakeManager.SendInput(key)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snakeInputResponse{
			Success: false,
			Error:   "no active game session",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snakeInputResponse{
		Success: true,
		State:   string(state),
	})
}

func (s *Server) handleSnakeStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
		return
	}

	s.snakeManager.StopSession()
	level.Info(s.logger).Log("msg", "Snake game stopped")
	writeJSON(w, apiResponse{Success: true})
}

func (s *Server) handleSnakeStatus(w http.ResponseWriter, r *http.Request) {
	state, active := s.snakeManager.GetState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snakeInputResponse{
		Success: active,
		State:   string(state),
	})
}

func writeGIFError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(apiResponse{Success: false, Error: message})
}

func (s *Server) handleTextPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeGIFError(w, "method not allowed")
		return
	}

	textMsg := r.URL.Query().Get("text")
	if textMsg == "" {
		writeGIFError(w, "text parameter required")
		return
	}

	animation := r.URL.Query().Get("animation")
	if animation == "" {
		animation = "none"
	}

	colorName := r.URL.Query().Get("color")
	if colorName == "" {
		colorName = "white"
	}

	colorName = strings.ToLower(strings.TrimSpace(colorName))
	color, ok := graphic.ColorPalette[colorName]
	if !ok {
		writeGIFError(w, fmt.Sprintf("unknown color: %s", colorName))
		return
	}

	msg := strings.ToUpper(textMsg)

	lines := text.WrapText(msg)
	blockHeight := text.TextBlockHeight(lines)
	if blockHeight > graphic.DisplayHeight {
		writeGIFError(w, fmt.Sprintf("text too long: wrapped to %d lines (%d pixels, max %d)", len(lines), blockHeight, graphic.DisplayHeight))
		return
	}

	opts := text.DefaultAnimationOptions()
	opts.TextOptions.TextColor = color
	opts.TextOptions.ShadowColor = graphic.ShadowFor(color)

	img, errMsg := text.GenerateAnimation(animation, msg, opts)
	if errMsg != "" {
		writeGIFError(w, errMsg)
		return
	}

	if parseMirrored(r) {
		img = img.Mirror()
	}

	img = img.AdjustBrightness(parseBrightness(r))

	var gifBytes []byte
	var err error
	if img.Type == graphic.ImageTypeStatic {
		rawBytes, err := img.RawBytes()
		if err != nil {
			writeGIFError(w, err.Error())
			return
		}
		paletted := graphic.RGBToPaletted(rawBytes)
		g := &gif.GIF{
			Image:     []*image.Paletted{paletted},
			Delay:     []int{0},
			LoopCount: 0,
		}
		var buf bytes.Buffer
		if err := gif.EncodeAll(&buf, g); err != nil {
			writeGIFError(w, err.Error())
			return
		}
		gifBytes = buf.Bytes()
	} else {
		gifBytes, err = img.GIFBytes()
		if err != nil {
			writeGIFError(w, err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(gifBytes)
}

func (s *Server) handleEmojiPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeGIFError(w, "method not allowed")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeGIFError(w, fmt.Sprintf("name parameter required (available: %s)", strings.Join(emoji.Names(), ", ")))
		return
	}

	img, err := emoji.Generate(name)
	if err != nil {
		writeGIFError(w, err.Error())
		return
	}

	if parseMirrored(r) {
		img = img.Mirror()
	}

	img = img.AdjustBrightness(parseBrightness(r))

	gifBytes, err := img.GIFBytes()
	if err != nil {
		writeGIFError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(gifBytes)
}

func (s *Server) handleGrotPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeGIFError(w, "method not allowed")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeGIFError(w, fmt.Sprintf("name parameter required (available: %s)", strings.Join(grot.Names(), ", ")))
		return
	}

	img, err := grot.Generate(name)
	if err != nil {
		writeGIFError(w, err.Error())
		return
	}

	if parseMirrored(r) {
		img = img.Mirror()
	}

	img = img.AdjustBrightness(parseBrightness(r))

	gifBytes, err := img.GIFBytes()
	if err != nil {
		writeGIFError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(gifBytes)
}

func (s *Server) handleFirePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeGIFError(w, "method not allowed")
		return
	}

	gifData := fire.GenerateGIF()

	g, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		writeGIFError(w, "failed to decode GIF: "+err.Error())
		return
	}
	if parseMirrored(r) {
		g = graphic.MirrorGIFHorizontal(g)
	}
	g = graphic.AdjustBrightnessGIF(g, parseBrightness(r))
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		writeGIFError(w, "failed to re-encode GIF: "+err.Error())
		return
	}
	gifData = buf.Bytes()

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(gifData)
}
