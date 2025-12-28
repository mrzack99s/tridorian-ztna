package common

import (
	"html/template"
	"net/http"
)

const blockedPageHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #f8f9fa;
            --card-bg: #ffffff;
            --text-primary: #202124;
            --text-secondary: #5f6368;
            --primary: #1a73e8;
            --error: #d93025;
            --border-color: #dadce0;
        }
        body {
            margin: 0;
            padding: 0;
            font-family: 'Google Sans', 'Roboto', Helvetica, Arial, sans-serif;
            background-color: var(--bg-color);
            height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: var(--text-primary);
        }
        .container {
            width: 100%;
            max-width: 440px;
            padding: 1rem;
            box-sizing: border-box;
        }
        .card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 48px 40px 36px;
            text-align: center;
            /* Minimal shadow typical of Google login boxes */
            box-shadow: 0 1px 2px 0 rgba(60,64,67,0.3), 0 2px 6px 2px rgba(60,64,67,0.15); 
        }
        .icon-wrapper {
            margin-bottom: 1rem;
        }
        .icon {
            color: var(--error);
            width: 48px;
            height: 48px;
        }
        h1 {
            font-family: 'Google Sans', 'Roboto', sans-serif;
            font-size: 24px;
            font-weight: 400;
            line-height: 1.3333;
            margin: 0 0 12px;
            padding-bottom: 0;
            padding-top: 0;
            color: var(--text-primary);
        }
        p {
            color: var(--text-secondary);
            font-size: 14px;
            line-height: 1.5;
            margin: 0 0 24px;
            letter-spacing: 0.2px;
        }
        .error-details {
            margin-top: 24px;
            padding: 12px;
            background-color: #fce8e6;
            border-radius: 4px;
            color: #c5221f;
            font-family: 'Roboto Mono', monospace;
            font-size: 12px;
            word-break: break-all;
            text-align: left;
            border: 1px solid #fad2cf;
        }
        .divider {
            height: 1px;
            background-color: var(--border-color);
            margin: 24px 0;
        }
        .btn {
            background-color: var(--primary);
            color: #fff;
            padding: 10px 24px;
            border-radius: 4px;
            font-weight: 500;
            font-size: 14px;
            text-decoration: none;
            display: inline-block;
            transition: background-color 0.2s;
            border: none;
            cursor: pointer;
            font-family: 'Google Sans', 'Roboto', sans-serif;
        }
        .btn:hover {
            background-color: #174ea6;
            box-shadow: 0 1px 2px 0 rgba(60,64,67,.302), 0 1px 3px 1px rgba(60,64,67,.149);
        }
        /* Google Sans emulation if not available */
        @font-face {
            font-family: 'Google Sans';
            font-style: normal;
            font-weight: 400;
            src: local('Google Sans'), local('GoogleSans-Regular'), url(https://fonts.gstatic.com/s/productsans/v5/HYvgU2fE2nRJvZ5JFAumwegdm0LZdjqr5-oayXSOefg.woff2) format('woff2');
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="icon-wrapper">
                <svg class="icon" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                </svg>
            </div>
            <h1>{{ .Title }}</h1>
            <p>{{ .Message }}</p>
            
            {{ if .Error }}
            <div class="error-details">
                <strong>Error:</strong> {{ .Error }}
            </div>
            {{ end }}
        </div>
    </div>
</body>
</html>
`

// RenderErrorPage renders a beautiful HTML error page
func RenderErrorPage(w http.ResponseWriter, status int, title, message, errStr string) {
	tmpl, err := template.New("error").Parse(blockedPageHTML)
	if err != nil {
		http.Error(w, errStr, status)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	tmpl.Execute(w, map[string]string{
		"Title":   title,
		"Message": message,
		"Error":   errStr,
	})
}
