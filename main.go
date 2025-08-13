package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tempest Image Finder</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Inter', system-ui, sans-serif;
            background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #334155 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            color: #f1f5f9;
            animation: gradientShift 12s ease infinite;
        }
        
        @keyframes gradientShift {
            0%, 100% { background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #334155 100%); }
            50% { background: linear-gradient(135deg, #1e1b4b 0%, #312e81 50%, #1e293b 100%); }
        }
        
        .container {
            background: rgba(15, 23, 42, 0.95);
            backdrop-filter: blur(20px);
            border: 1px solid rgba(71, 85, 105, 0.3);
            border-radius: 24px;
            padding: 48px;
            box-shadow: 0 32px 64px rgba(0, 0, 0, 0.4);
            max-width: 580px;
            width: 100%;
            text-align: center;
            position: relative;
            overflow: hidden;
        }
        
        .container::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 4px;
            background: linear-gradient(90deg, #3b82f6, #8b5cf6, #06b6d4);
            animation: shimmer 3s ease-in-out infinite;
        }
        
        @keyframes shimmer {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.7; }
        }
        
        h1 {
            color: #f1f5f9;
            margin-bottom: 12px;
            font-size: 2.5rem;
            font-weight: 700;
            letter-spacing: -0.025em;
            background: linear-gradient(135deg, #3b82f6, #8b5cf6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .subtitle {
            color: #94a3b8;
            margin-bottom: 20px;
            font-size: 1.1rem;
            font-weight: 400;
        }
        
        .disclaimer {
            background: rgba(34, 197, 94, 0.1);
            border: 1px solid rgba(34, 197, 94, 0.3);
            border-radius: 12px;
            padding: 16px;
            margin-bottom: 20px;
            color: #4ade80;
            font-size: 0.85rem;
            line-height: 1.4;
            text-align: left;
        }
        
        .disclaimer-title {
            font-weight: 600;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .warning {
            background: rgba(245, 158, 11, 0.1);
            border: 1px solid rgba(245, 158, 11, 0.3);
            border-radius: 12px;
            padding: 16px;
            margin-bottom: 32px;
            color: #fbbf24;
            font-size: 0.9rem;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .form-group {
            margin-bottom: 32px;
            text-align: left;
            position: relative;
        }
        
        label {
            display: block;
            margin-bottom: 12px;
            color: #e2e8f0;
            font-weight: 600;
            font-size: 0.95rem;
        }
        
        .input-container {
            position: relative;
        }
        
        input {
            width: 100%;
            padding: 18px 24px;
            border: 2px solid #374151;
            border-radius: 16px;
            font-size: 16px;
            transition: all 0.3s ease;
            background: rgba(30, 41, 59, 0.8);
            color: #f1f5f9;
            font-family: 'JetBrains Mono', 'Courier New', monospace;
            font-weight: 500;
        }
        
        input:focus {
            outline: none;
            border-color: #3b82f6;
            box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.2);
            transform: translateY(-2px);
            background: rgba(30, 41, 59, 1);
        }
        
        input::placeholder {
            color: #6b7280;
            font-weight: 400;
        }
        
        .suggestions {
            position: absolute;
            top: 100%;
            left: 0;
            right: 0;
            background: rgba(15, 23, 42, 0.98);
            border: 2px solid #374151;
            border-top: none;
            border-radius: 0 0 16px 16px;
            max-height: 200px;
            overflow-y: auto;
            z-index: 10;
            display: none;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
        }
        
        .suggestion-item {
            padding: 14px 24px;
            cursor: pointer;
            font-family: 'JetBrains Mono', 'Courier New', monospace;
            font-size: 14px;
            color: #e2e8f0;
            border-bottom: 1px solid #374151;
            transition: all 0.2s ease;
            font-weight: 500;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .suggestion-item:hover {
            background: linear-gradient(135deg, #3b82f6, #8b5cf6);
            color: white;
            transform: translateX(4px);
        }
        
        .suggestion-item:last-child {
            border-bottom: none;
        }
        
        .suggestion-remove {
            color: #f87171;
            font-size: 12px;
            padding: 2px 6px;
            border-radius: 4px;
            background: rgba(248, 113, 113, 0.1);
            opacity: 0;
            transition: opacity 0.2s ease;
        }
        
        .suggestion-item:hover .suggestion-remove {
            opacity: 1;
        }
        
        button {
            background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
            color: white;
            border: none;
            padding: 18px 36px;
            border-radius: 16px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-top: 24px;
            position: relative;
            overflow: hidden;
            min-width: 160px;
        }
        
        button::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
            transition: left 0.5s;
        }
        
        button:hover {
            transform: translateY(-3px);
            box-shadow: 0 20px 40px rgba(59, 130, 246, 0.4);
        }
        
        button:hover::before {
            left: 100%;
        }
        
        button:active {
            transform: translateY(-1px);
        }
        
        button:disabled {
            opacity: 0.7;
            cursor: not-allowed;
            transform: none;
        }
        

        .image-container {
            margin-top: 40px;
            border-radius: 20px;
            overflow: hidden;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
            display: none;
            border: 3px solid rgba(71, 85, 105, 0.5);
            position: relative;
            /* Prevent iOS context menu from interfering */
            -webkit-touch-callout: default;
            -webkit-user-select: none;
            -moz-user-select: none;
            -ms-user-select: none;
            user-select: none;
        }
        
        .image-container::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: linear-gradient(45deg, transparent 30%, rgba(255,255,255,0.05) 50%, transparent 70%);
            pointer-events: none;
            z-index: 1;
        }
        
        .image-container img {
            width: 100%;
            height: auto;
            display: block;
            transition: transform 0.3s ease;
            /* Enable right-click/long-press save on mobile */
            -webkit-touch-callout: default;
            -webkit-user-select: auto;
            -moz-user-select: auto;
            -ms-user-select: auto;
            user-select: auto;
            /* Prevent drag on desktop but allow save on mobile */
            -webkit-user-drag: none;
            -khtml-user-drag: none;
            -moz-user-drag: none;
            -o-user-drag: none;
            user-drag: none;
        }
        
        .image-container:hover img {
            transform: scale(1.02);
        }
        
        .image-download-hint {
            margin-top: 16px;
            color: #94a3b8;
            font-size: 0.85rem;
            padding: 12px;
            background: rgba(71, 85, 105, 0.2);
            border-radius: 12px;
            font-style: italic;
        }
        
        .loading {
            display: none;
            margin-top: 32px;
            color: #3b82f6;
            font-weight: 500;
            align-items: center;
            justify-content: center;
            flex-direction: column;
            gap: 20px;
        }
        
        .loading-spinner {
            width: 48px;
            height: 48px;
            border: 4px solid rgba(59, 130, 246, 0.2);
            border-top: 4px solid #3b82f6;
            border-radius: 50%;
            animation: spin 1s cubic-bezier(0.68, -0.55, 0.265, 1.55) infinite;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .loading-text {
            font-size: 1.1rem;
            font-weight: 500;
        }
        
        .loading-dots {
            display: flex;
            gap: 6px;
        }
        
        .loading-dot {
            width: 10px;
            height: 10px;
            background: #3b82f6;
            border-radius: 50%;
            animation: bounce 1.4s ease-in-out infinite both;
        }
        
        .loading-dot:nth-child(1) { animation-delay: -0.32s; }
        .loading-dot:nth-child(2) { animation-delay: -0.16s; }
        
        @keyframes bounce {
            0%, 80%, 100% {
                transform: scale(0);
            }
            40% {
                transform: scale(1);
            }
        }
        
        .error {
            color: #f87171;
            margin-top: 24px;
            display: none;
            padding: 20px;
            background: rgba(248, 113, 113, 0.1);
            border: 2px solid rgba(248, 113, 113, 0.3);
            border-radius: 16px;
            font-size: 15px;
            font-weight: 500;
            backdrop-filter: blur(10px);
        }
        
        .status-info {
            margin-top: 24px;
            padding: 16px;
            background: rgba(34, 197, 94, 0.1);
            border: 2px solid rgba(34, 197, 94, 0.3);
            border-radius: 16px;
            font-size: 14px;
            color: #4ade80;
            display: none;
            font-weight: 500;
            backdrop-filter: blur(10px);
        }

        @media (max-width: 640px) {
            .container {
                padding: 32px 24px;
                margin: 10px;
            }
            
            h1 {
                font-size: 2rem;
            }
            
            input, button {
                padding: 16px 20px;
            }
            
            .image-download-hint {
                font-size: 0.8rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Image Finder</h1>
        <p class="subtitle">Find and retrieve your images instantly ✨</p>
        
        <div class="disclaimer">
            <div class="disclaimer-title">
                🔒 Privacy Notice
            </div>
            <div>Images are fetched directly from Tempest and displayed in your browser only. No images are stored on our servers or visible to anyone else. Your image searches are completely private.</div>
        </div>
        
        <div class="warning">
            ⚠️ Large images may take some time to process and load
        </div>
        
        <form id="photoForm">
            <div class="form-group">
                <label for="photoId">Image ID</label>
                <div class="input-container">
                    <input 
                        type="text" 
                        id="photoId" 
                        name="photoId" 
                        value="89715C5328" 
                        placeholder="Enter your image ID..."
                        autocomplete="off"
                        required
                    >
                    <div class="suggestions" id="suggestions"></div>
                </div>
            </div>
            
            <button type="submit" id="submitBtn">
                <span class="btn-text">Get Image</span>
            </button>

        </form>
        
        <div class="loading" id="loading">
            <div class="loading-spinner"></div>
            <div class="loading-text">Finding your image...</div>
            <div class="loading-dots">
                <div class="loading-dot"></div>
                <div class="loading-dot"></div>
                <div class="loading-dot"></div>
            </div>
        </div>
        
        <div class="error" id="error"></div>
        <div class="status-info" id="statusInfo"></div>
        
        <div class="image-container" id="imageContainer">
            <img id="photo" src="" alt="Retrieved Image">
            <div class="image-download-hint">
                📱 On mobile: Long press the image to save to your camera roll
            </div>
        </div>
    </div>

    <script>
        // Note: This uses in-memory storage for demo purposes
        // In a real environment, replace with localStorage for persistence
        let recentIds = ['89715C5328'];
        
        // Storage functions - replace with localStorage in production
        function loadRecentIds() {
            try {
                // For production use: 
                // const stored = localStorage.getItem('tempest-recent-ids');
                // return stored ? JSON.parse(stored) : ['89715C5328'];
                return recentIds;
            } catch (e) {
                console.warn('Could not load recent IDs:', e);
                return ['89715C5328'];
            }
        }
        
        function saveRecentIds(ids) {
            try {
                // For production use:
                // localStorage.setItem('tempest-recent-ids', JSON.stringify(ids));
                recentIds = ids;
            } catch (e) {
                console.warn('Could not save recent IDs:', e);
            }
        }
        
        // Initialize recent IDs
        recentIds = loadRecentIds();
        
        const photoIdInput = document.getElementById('photoId');
        const suggestionsDiv = document.getElementById('suggestions');
        
        photoIdInput.addEventListener('focus', showSuggestions);
        photoIdInput.addEventListener('input', filterSuggestions);
        

        
        document.addEventListener('click', function(e) {
            if (!e.target.closest('.input-container')) {
                hideSuggestions();
            }
        });
        
        // Prevent page refresh on image long-press (iOS Safari fix)
        document.addEventListener('contextmenu', function(e) {
            if (e.target.tagName === 'IMG') {
                e.stopPropagation();
                // Allow the context menu for image saving
                return true;
            }
        });
        
        // Prevent pull-to-refresh when interacting with images
        let touchStartY = 0;
        document.addEventListener('touchstart', function(e) {
            if (e.target.closest('.image-container')) {
                touchStartY = e.touches[0].clientY;
            }
        });
        
        document.addEventListener('touchmove', function(e) {
            if (e.target.closest('.image-container')) {
                const touchY = e.touches[0].clientY;
                const touchDelta = touchY - touchStartY;
                
                // Prevent pull-to-refresh when scrolling up from image
                if (touchDelta > 0 && window.scrollY === 0) {
                    e.preventDefault();
                }
            }
        });
        
        function showSuggestions() {
            if (recentIds.length > 0) {
                updateSuggestionsList(recentIds);
                suggestionsDiv.style.display = 'block';
            }
        }
        
        function filterSuggestions() {
            const value = photoIdInput.value.toLowerCase();
            if (value === '') {
                updateSuggestionsList(recentIds);
            } else {
                const filtered = recentIds.filter(id => 
                    id.toLowerCase().includes(value)
                );
                updateSuggestionsList(filtered);
            }
            
            if (suggestionsDiv.children.length > 0) {
                suggestionsDiv.style.display = 'block';
            } else {
                suggestionsDiv.style.display = 'none';
            }
        }
        
        function updateSuggestionsList(ids) {
            suggestionsDiv.innerHTML = '';
            ids.forEach(id => {
                const div = document.createElement('div');
                div.className = 'suggestion-item';
                
                const idSpan = document.createElement('span');
                idSpan.textContent = id;
                
                const removeBtn = document.createElement('span');
                removeBtn.className = 'suggestion-remove';
                removeBtn.textContent = '×';
                removeBtn.title = 'Remove from history';
                
                div.appendChild(idSpan);
                div.appendChild(removeBtn);
                
                idSpan.addEventListener('click', function() {
                    photoIdInput.value = id;
                    hideSuggestions();
                });
                
                removeBtn.addEventListener('click', function(e) {
                    e.stopPropagation();
                    removeFromRecentIds(id);
                    filterSuggestions();
                });
                
                suggestionsDiv.appendChild(div);
            });
        }
        
        function hideSuggestions() {
            suggestionsDiv.style.display = 'none';
        }
        
        function addToRecentIds(id) {
            recentIds = recentIds.filter(existingId => existingId !== id);
            recentIds.unshift(id);
            recentIds = recentIds.slice(0, 10);
            saveRecentIds(recentIds);
        }
        
        function removeFromRecentIds(id) {
            recentIds = recentIds.filter(existingId => existingId !== id);
            saveRecentIds(recentIds);

        }
        
        
        document.getElementById('photoForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const photoId = document.getElementById('photoId').value.trim();
            const loading = document.getElementById('loading');
            const error = document.getElementById('error');
            const statusInfo = document.getElementById('statusInfo');
            const imageContainer = document.getElementById('imageContainer');
            const photo = document.getElementById('photo');
            const submitBtn = document.getElementById('submitBtn');
            
            if (!photoId) {
                error.textContent = '🤔 Please enter an image ID first!';
                error.style.display = 'block';
                return;
            }
            
            loading.style.display = 'flex';
            error.style.display = 'none';
            statusInfo.style.display = 'none';
            imageContainer.style.display = 'none';
            submitBtn.disabled = true;
            hideSuggestions();
            
            const startTime = Date.now();
            
            try {
                const response = await fetch(` + "`/fetch-photo?id=${encodeURIComponent(photoId)}`" + `);
                
                const contentType = response.headers.get('content-type');
                
                if (!response.ok) {
                    let errorMessage = ` + "`Failed to fetch image (${response.status})`" + `;
                    
                    if (contentType && contentType.includes('application/json')) {
                        const errorData = await response.json();
                        errorMessage = errorData.error || errorMessage;
                        
                        if (response.status === 404) {
                            errorMessage = ` + "`🔍 Image '${photoId}' not found. Double-check your ID!`" + `;
                        } else if (response.status === 403) {
                            errorMessage = ` + "`🔒 Access denied for image '${photoId}'. You might not have permission.`" + `;
                        } else if (response.status === 500) {
                            errorMessage = ` + "`⚠️ ${errorData.details || 'Server error occurred while fetching the image'}`" + `;
                        } else if (response.status === 408) {
                            errorMessage = ` + "`⏱️ Request timed out. The image may be too large or the server is busy.`" + `;
                        }
                    }
                    
                    throw new Error(errorMessage);
                }
                
                const blob = await response.blob();
                const imageUrl = URL.createObjectURL(blob);
                const loadTime = ((Date.now() - startTime) / 1000).toFixed(2);
                
                photo.src = imageUrl;
                photo.onload = function() {
                    loading.style.display = 'none';
                    statusInfo.textContent = ` + "`✅ Image loaded successfully in ${loadTime}s`" + `;
                    statusInfo.style.display = 'block';
                    imageContainer.style.display = 'block';
                    submitBtn.disabled = false;
                    
                    addToRecentIds(photoId);
                    
                    setTimeout(() => {
                        statusInfo.style.display = 'none';
                    }, 4000);
                };
                
            } catch (err) {
                loading.style.display = 'none';
                error.textContent = err.message;
                error.style.display = 'block';
                statusInfo.style.display = 'none';
                submitBtn.disabled = false;
            }
        });
    </script>
</body>
</html>
`

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"status"`
}

func sendJSONError(w http.ResponseWriter, message string, details string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:   message,
		Details: details,
		Status:  statusCode,
	}

	json.NewEncoder(w).Encode(errorResp)
}

func main() {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] GET / - Client: %s\n", time.Now().Format("15:04:05"), r.RemoteAddr)
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/fetch-photo", func(w http.ResponseWriter, r *http.Request) {
		photoId := r.URL.Query().Get("id")
		clientIP := r.RemoteAddr

		fmt.Printf("[%s] POST /fetch-photo - Client: %s - ID: %s\n", time.Now().Format("15:04:05"), clientIP, photoId)

		if photoId == "" {
			fmt.Printf("[%s] ERROR: Missing photo ID from %s\n", time.Now().Format("15:04:05"), clientIP)
			sendJSONError(w, "Image ID required", "Please provide a valid image identifier", http.StatusBadRequest)
			return
		}

		apiURL := fmt.Sprintf("https://us-central1-htempest-preproduction-prod.cloudfunctions.net/ImageApiProxy/image/%s/preview/?exifrotate=1&MaxSize=9999&ProofWatermark=FALSE&source=G&WithCrop=TRUE", photoId)

		fmt.Printf("[%s] Requesting Tempest API for ID: %s\n", time.Now().Format("15:04:05"), photoId)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			fmt.Printf("[%s] ERROR: Failed to create request for ID %s: %v\n", time.Now().Format("15:04:05"), photoId, err)
			sendJSONError(w, "Request creation failed", fmt.Sprintf("Unable to create API request: %v", err), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("[%s] TIMEOUT: Tempest API request timed out for ID %s after 20s\n", time.Now().Format("15:04:05"), photoId)
				sendJSONError(w, "Request timeout", "The image request took too long to process (>20s). The image may be very large.", http.StatusRequestTimeout)
				return
			}
			fmt.Printf("[%s] ERROR: Tempest API connection failed for ID %s: %v\n", time.Now().Format("15:04:05"), photoId, err)
			sendJSONError(w, "Connection failed", fmt.Sprintf("Unable to connect to Tempest API: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("[%s] Tempest API response for ID %s: %d %s\n", time.Now().Format("15:04:05"), photoId, resp.StatusCode, resp.Status)

		if resp.StatusCode == 204 {
			fmt.Printf("[%s] NOT FOUND: Image ID %s not found in Tempest\n", time.Now().Format("15:04:05"), photoId)
			sendJSONError(w, "Image not found", fmt.Sprintf("The image ID '%s' was not found in the Tempest system", photoId), http.StatusNotFound)
			return
		}
		switch resp.StatusCode {
		case http.StatusOK:
			contentLength := resp.Header.Get("Content-Length")
			fmt.Printf("[%s] SUCCESS: Serving image %s (Content-Length: %s) to %s\n", time.Now().Format("15:04:05"), photoId, contentLength, clientIP)
			w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			w.Header().Set("Cache-Control", "public, max-age=3600")
			io.Copy(w, resp.Body)

		case http.StatusForbidden:
			fmt.Printf("[%s] FORBIDDEN: Access denied for image ID %s\n", time.Now().Format("15:04:05"), photoId)
			sendJSONError(w, "Access denied", fmt.Sprintf("You don't have permission to access image '%s'", photoId), http.StatusForbidden)

		case http.StatusUnauthorized:
			fmt.Printf("[%s] UNAUTHORIZED: Authentication required for image ID %s\n", time.Now().Format("15:04:05"), photoId)
			sendJSONError(w, "Authentication required", "The request requires valid authentication credentials", http.StatusUnauthorized)

		case http.StatusInternalServerError:
			fmt.Printf("[%s] SERVER ERROR: Tempest API internal error for image ID %s\n", time.Now().Format("15:04:05"), photoId)
			sendJSONError(w, "Tempest API error", "The upstream image service is currently experiencing issues", http.StatusInternalServerError)

		case http.StatusServiceUnavailable:
			fmt.Printf("[%s] UNAVAILABLE: Tempest API service unavailable for image ID %s\n", time.Now().Format("15:04:05"), photoId)
			sendJSONError(w, "Service unavailable", "The Tempest API is temporarily unavailable. Please try again later.", http.StatusServiceUnavailable)

		default:
			fmt.Printf("[%s] UNEXPECTED: Tempest API returned %d for image ID %s\n", time.Now().Format("15:04:05"), resp.StatusCode, photoId)
			sendJSONError(w, "Unexpected error", fmt.Sprintf("Tempest API returned status %d", resp.StatusCode), resp.StatusCode)
		}
	})

	fmt.Printf("[%s] 🚀 Image Finder starting on http://localhost:8080\n", time.Now().Format("15:04:05"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
