package common

import (
	"fmt"
	"html/template"
	"net/http"
)

const blockedPageHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }} - Tridorian ZTNA</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #4f46e5;
            --danger: #ef4444;
            --bg-page: #f8fafc;
            --card-bg: #ffffff;
            --text-main: #1e293b;
            --text-dim: #64748b;
            --border: #e2e8f0;
            --detail-bg: #f1f5f9;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Outfit', sans-serif;
            background-color: var(--bg-page);
            background-image: 
                radial-gradient(at 0% 0%, rgba(79, 70, 229, 0.05) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(239, 68, 68, 0.03) 0px, transparent 50%);
            height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: var(--text-main);
        }

        .container {
            width: 100%;
            max-width: 440px;
            padding: 24px;
            perspective: 1000px;
        }

        .card {
            background: var(--card-bg);
            border: 1px solid var(--border);
            border-radius: 32px;
            padding: 56px 40px;
            text-align: center;
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.05), 0 10px 10px -5px rgba(0, 0, 0, 0.02);
            animation: slideUp 0.7s cubic-bezier(0.16, 1, 0.3, 1);
        }

        @keyframes slideUp {
            from { opacity: 0; transform: translateY(30px) scale(0.98); }
            to { opacity: 1; transform: translateY(0) scale(1); }
        }

        .icon-box {
            width: 72px;
            height: 72px;
            background: #fff1f2;
            border-radius: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 32px;
            color: var(--danger);
            border: 1px solid #ffe4e6;
        }

        h1 {
            font-size: 26px;
            font-weight: 600;
            margin-bottom: 16px;
            letter-spacing: -0.02em;
            color: var(--text-main);
        }

        .message {
            color: var(--text-dim);
            font-size: 15px;
            line-height: 1.6;
            margin-bottom: 32px;
        }

        .technical-details {
            background: var(--detail-bg);
            border-radius: 20px;
            padding: 24px;
            text-align: left;
        }

        .detail-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 12px;
            font-size: 13px;
        }

        .detail-row:last-child {
            margin-bottom: 0;
        }

        .label {
            color: var(--text-dim);
            font-weight: 400;
        }

        .value {
            color: var(--text-main);
            font-weight: 600;
            font-family: 'ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', monospace;
        }

        .error-msg {
            margin-top: 16px;
            padding-top: 16px;
            border-top: 1px solid rgba(0,0,0,0.05);
            color: var(--danger);
            font-size: 12px;
            opacity: 0.8;
            word-break: break-all;
        }

        .footer {
            margin-top: 40px;
            font-size: 12px;
            color: var(--text-dim);
            letter-spacing: 0.05em;
            text-transform: uppercase;
            font-weight: 500;
        }

        /* Abstract soft shadows */
        .card::after {
            content: '';
            position: absolute;
            top: 10px; left: 10px; right: 10px; bottom: -10px;
            background: var(--primary);
            opacity: 0.03;
            filter: blur(40px);
            z-index: -1;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="icon-box">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
            </div>
            
            <h1>{{ .Title }}</h1>
            <p class="message">{{ .Message }}</p>

            <div class="technical-details">
                <div class="detail-row">
                    <span class="label">IP Address</span>
                    <span class="value">{{ if .IP }}{{ .IP }}{{ else }}—{{ end }}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Location</span>
                    <span class="value">{{ if .Country }}{{ .Country }}{{ else }}Unknown{{ end }}</span>
                </div>
                {{ if .Error }}
                <div class="error-msg">
                    {{ .Error }}
                </div>
                {{ end }}
            </div>
            
            <div class="footer">
                Tridorian ZTNA Security
            </div>
        </div>
    </div>
</body>
</html>
`

// RenderErrorPage renders a beautiful HTML error page
func RenderErrorPage(w http.ResponseWriter, status int, title, message, errStr string, ip string, country string) {
	tmpl, err := template.New("error").Parse(blockedPageHTML)
	if err != nil {
		if ip != "" || country != "" {
			errStr = errStr + fmt.Sprintf(" (IP: %s, Location: %s)", ip, country)
		}
		http.Error(w, errStr, status)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	tmpl.Execute(w, map[string]string{
		"Title":   title,
		"Message": message,
		"Error":   errStr,
		"IP":      ip,
		"Country": country,
	})
}
