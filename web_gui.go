package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	guiCrawler *MainCrawler
	guiMutex   sync.Mutex
	guiLogs    []string
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Pavuk Crawler GUI</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f4f4f9; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        h1 { color: #333; }
        label { display: block; margin-top: 10px; font-weight: bold; }
        input[type="text"], input[type="number"], textarea { width: 100%; padding: 8px; margin-top: 5px; box-sizing: border-box; }
        button { margin-top: 20px; padding: 10px 20px; background-color: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button.stop { background-color: #dc3545; }
        .log-box { margin-top: 20px; background: #222; color: #0f0; padding: 10px; height: 300px; overflow-y: scroll; font-family: monospace; border-radius: 4px; }
        .status { margin-top: 20px; font-size: 1.2em; }
        .stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin-top: 20px; margin-bottom: 20px;}
        .stat-box { background: #e9ecef; padding: 15px; border-radius: 8px; text-align: center; }
        .stat-val { font-size: 1.5em; font-weight: bold; color: #007bff; }
    </style>
</head>
<body>
<div class="container">
    <h1>Pavuk Web Crawler</h1>

    <div class="status">Status: <strong id="runStatus">{{.Status}}</strong></div>

    <div class="stats-grid">
        <div class="stat-box">Domains<div class="stat-val" id="statDomains">0 / 0</div></div>
        <div class="stat-box">Pages Crawled<div class="stat-val" id="statPages">0</div></div>
        <div class="stat-box">Target Hits<div class="stat-val" id="statFindings">0</div></div>
        <div class="stat-box">Errors<div class="stat-val" id="statErrors">0</div></div>
    </div>

    <form id="crawlerForm" method="POST" action="/start">
        <label>Domains (one per line):</label>
        <textarea name="domains" rows="5">example.com</textarea>

        <label>Proxies (one per line, e.g. http://user:pass@ip:port):</label>
        <textarea name="proxies" rows="3"></textarea>

        <label>Workers per Domain:</label>
        <input type="number" name="workers" value="1">

        <label>Max Total Workers:</label>
        <input type="number" name="maxWorkers" value="100">

        <label>Max Pages per Domain:</label>
        <input type="number" name="maxPages" value="10">

        <button type="submit">Start Crawler</button>
    </form>

    <form method="POST" action="/stop" style="display:inline;">
        <button class="stop" type="submit">Stop Crawler</button>
    </form>

    <div class="log-box" id="logBox">
        {{range .Logs}}
            <div>{{.}}</div>
        {{end}}
    </div>
</div>
<script>
    setInterval(function(){
        fetch('/api/data').then(r => r.json()).then(data => {
            document.getElementById('runStatus').textContent = data.Status;
            document.getElementById('statDomains').textContent = data.DomainsDone + " / " + data.DomainsTotal;
            document.getElementById('statPages').textContent = data.Pages;
            document.getElementById('statFindings').textContent = data.Findings;
            document.getElementById('statErrors').textContent = data.Errors;

            const box = document.getElementById('logBox');
            box.innerHTML = '';
            data.Logs.forEach(line => {
                const d = document.createElement('div');
                d.textContent = line;
                box.appendChild(d);
            });
            box.scrollTop = box.scrollHeight;
        });
    }, 1500);
</script>
</body>
</html>
`

func webGuiMain() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		guiMutex.Lock()
		status := "Idle"
		if guiCrawler != nil {
			status = "Running"
		}
		logsCopy := make([]string, len(guiLogs))
		copy(logsCopy, guiLogs)
		guiMutex.Unlock()

		tmpl, _ := template.New("web").Parse(htmlTemplate)
		tmpl.Execute(w, map[string]interface{}{
			"Status": status,
			"Logs":   logsCopy,
		})
	})

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		domains := strings.Split(r.FormValue("domains"), "\n")
		proxies := strings.Split(r.FormValue("proxies"), "\n")
		workers := parseInt(r.FormValue("workers"), 1)
		maxWorkers := parseInt(r.FormValue("maxWorkers"), 100)
		maxPages := parseInt(r.FormValue("maxPages"), 10)

		domainFile := "gui_domains.txt"
		proxyFile := "gui_proxies.txt"

		saveLines(domainFile, domains)
		saveLines(proxyFile, proxies)

		guiMutex.Lock()
		if guiCrawler != nil {
			guiMutex.Unlock()
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		guiLogs = append(guiLogs, "Starting crawler via GUI...")
		guiMutex.Unlock()

		go func() {
			config := DefaultConfig()
			config.WorkersPerDomain = workers
			config.MaxTotalWorkers = maxWorkers
			config.MaxPagesPerDomain = maxPages
			config.ProxiesFile = proxyFile

			if len(proxies) > 0 && proxies[0] != "" {
				config.Proxies, _ = loadLines(proxyFile)
			}

			c, err := NewMainCrawler(config)
			if err != nil {
				AddLog(fmt.Sprintf("Failed to create crawler: %v", err))

				guiMutex.Lock()
				guiCrawler = nil
				guiMutex.Unlock()
				return
			}

			guiMutex.Lock()
			guiCrawler = c
			guiMutex.Unlock()

			if err := c.Run(domainFile); err != nil {
				AddLog(fmt.Sprintf("Crawler failed: %v", err))
			}

			guiMutex.Lock()
			guiCrawler = nil
			guiLogs = append(guiLogs, "Crawler finished.")
			guiMutex.Unlock()
		}()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		guiMutex.Lock()
		if guiCrawler != nil {
			guiCrawler.Shutdown()
			guiLogs = append(guiLogs, "Stop signal sent to crawler.")
		}
		guiMutex.Unlock()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		guiMutex.Lock()
		status := "Idle"
		var domainsTotal, domainsDone, pages, findings, errors int64

		if guiCrawler != nil {
			status = "Running"
			domainsTotal = atomic.LoadInt64(&guiCrawler.domainsTotal)
			domainsDone = atomic.LoadInt64(&guiCrawler.domainsDone)
			pages = atomic.LoadInt64(&guiCrawler.pagesTotal)
			findings = atomic.LoadInt64(&guiCrawler.findingsTotal)
			errors = atomic.LoadInt64(&guiCrawler.errorsTotal)
		}

		logsCopy := make([]string, len(guiLogs))
		copy(logsCopy, guiLogs)
		guiMutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status":       status,
			"DomainsTotal": domainsTotal,
			"DomainsDone":  domainsDone,
			"Pages":        pages,
			"Findings":     findings,
			"Errors":       errors,
			"Logs":         logsCopy,
		})
	})

	log.Println("Starting Web GUI on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func AddLog(msg string) {
	log.Println(msg)
	guiMutex.Lock()
	guiLogs = append(guiLogs, msg)
	if len(guiLogs) > 100 {
		guiLogs = guiLogs[1:]
	}
	guiMutex.Unlock()
}

func parseInt(s string, def int) int {
	var val int
	fmt.Sscanf(s, "%d", &val)
	if val <= 0 {
		return def
	}
	return val
}

func saveLines(path string, lines []string) {
	f, _ := os.Create(path)
	defer f.Close()
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			f.WriteString(l + "\n")
		}
	}
}
