#!/bin/bash
# Real-time Test Observability Dashboard
# Monitors test execution and provides live feedback

set -e

REPORT_DIR="test-reports/observability"
EVENT_FILE="$REPORT_DIR/test-events.jsonl"
HTML_FILE="$REPORT_DIR/realtime-dashboard.html"
PORT=8765

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
NC='\033[0m'

echo "=========================================="
echo "🔴 Real-time Test Observability Dashboard"
echo "=========================================="
echo ""

# Generate HTML Dashboard
cat > "$HTML_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Real-time Test Monitor</title>
    <meta http-equiv="refresh" content="2">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0d1117;
            color: #c9d1d9;
            padding: 20px;
        }

        .container {
            max-width: 1600px;
            margin: 0 auto;
        }

        .header {
            background: linear-gradient(135deg, #bd34fe 0%, #41d1ff 100%);
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 20px;
            text-align: center;
        }

        .header h1 {
            font-size: 36px;
            margin-bottom: 10px;
            color: white;
        }

        .status-bar {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }

        .status-card {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
        }

        .status-card.running {
            border-left: 4px solid #58a6ff;
            background: linear-gradient(135deg, #161b22 0%, #1c2533 100%);
        }

        .status-card.passed {
            border-left: 4px solid #3fb950;
        }

        .status-card.failed {
            border-left: 4px solid #f85149;
        }

        .status-value {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 5px;
        }

        .status-value.running { color: #58a6ff; }
        .status-value.passed { color: #3fb950; }
        .status-value.failed { color: #f85149; }
        .status-value.pending { color: #d29922; }

        .status-label {
            color: #8b949e;
            font-size: 14px;
            text-transform: uppercase;
        }

        .progress-section {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }

        .progress-bar {
            height: 30px;
            background: #21262d;
            border-radius: 15px;
            overflow: hidden;
            position: relative;
            margin-bottom: 10px;
        }

        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #3fb950 0%, #58a6ff 100%);
            transition: width 0.3s ease;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
        }

        .progress-info {
            display: flex;
            justify-content: space-between;
            color: #8b949e;
            font-size: 14px;
        }

        .specs-section {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }

        .specs-section h2 {
            margin-bottom: 15px;
            color: #58a6ff;
        }

        .spec-item {
            background: #0d1117;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 15px;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            gap: 15px;
        }

        .spec-status {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
            flex-shrink: 0;
        }

        .spec-status.running {
            background: #58a6ff;
            animation: pulse 1.5s ease-in-out infinite;
        }

        .spec-status.passed {
            background: #3fb950;
        }

        .spec-status.failed {
            background: #f85149;
        }

        .spec-status.pending {
            background: #8b949e;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .spec-info {
            flex: 1;
        }

        .spec-name {
            font-size: 16px;
            margin-bottom: 5px;
            color: #c9d1d9;
        }

        .spec-meta {
            font-size: 12px;
            color: #8b949e;
        }

        .spec-duration {
            font-size: 14px;
            color: #8b949e;
            font-family: 'Courier New', monospace;
        }

        .spec-error {
            margin-top: 10px;
            padding: 10px;
            background: #2d1517;
            border-left: 3px solid #f85149;
            border-radius: 4px;
            color: #ffa198;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }

        .live-indicator {
            position: fixed;
            top: 20px;
            right: 20px;
            background: #3fb950;
            color: white;
            padding: 10px 20px;
            border-radius: 20px;
            font-weight: bold;
            display: flex;
            align-items: center;
            gap: 10px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        }

        .live-dot {
            width: 10px;
            height: 10px;
            background: white;
            border-radius: 50%;
            animation: blink 1s ease-in-out infinite;
        }

        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.3; }
        }

        .timestamp {
            text-align: center;
            color: #8b949e;
            font-size: 12px;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="live-indicator">
        <div class="live-dot"></div>
        <span>LIVE</span>
    </div>

    <div class="container">
        <div class="header">
            <h1>🔴 Real-time Test Monitor</h1>
            <p>Live Test Execution Dashboard</p>
        </div>

        <div class="status-bar" id="statusBar">
            <!-- Will be populated by JavaScript -->
        </div>

        <div class="progress-section" id="progressSection">
            <!-- Will be populated by JavaScript -->
        </div>

        <div class="specs-section">
            <h2>📊 Test Execution Timeline</h2>
            <div id="specsList">
                <!-- Will be populated by JavaScript -->
            </div>
        </div>

        <div class="timestamp" id="timestamp"></div>
    </div>

    <script>
        // Parse JSONL event file and render dashboard
        function loadTestData() {
            // Read event file (simulated - in real implementation would use WebSocket or fetch)
            const eventData = `__EVENT_DATA__`;

            if (!eventData || eventData === '__EVENT_DATA__') {
                document.getElementById('statusBar').innerHTML = `
                    <div class="status-card">
                        <div class="status-value">0</div>
                        <div class="status-label">Waiting for tests...</div>
                    </div>
                `;
                return;
            }

            const events = eventData.trim().split('\\n').filter(line => line).map(line => {
                try {
                    return JSON.parse(line);
                } catch (e) {
                    return null;
                }
            }).filter(e => e);

            if (events.length === 0) {
                return;
            }

            // Get latest suite state
            const suiteEvents = events.filter(e => e.suite);
            const latestSuite = suiteEvents.length > 0 ? suiteEvents[suiteEvents.length - 1].suite : null;

            if (!latestSuite) {
                return;
            }

            renderStatusBar(latestSuite);
            renderProgress(latestSuite);

            // Get all specs
            const specEvents = events.filter(e => e.spec);
            renderSpecs(specEvents);

            document.getElementById('timestamp').textContent =
                'Last updated: ' + new Date().toLocaleTimeString();
        }

        function renderStatusBar(suite) {
            const progress = suite.total_specs > 0
                ? Math.round((suite.completed_specs / suite.total_specs) * 100)
                : 0;

            const html = `
                <div class="status-card running">
                    <div class="status-value running">${progress}%</div>
                    <div class="status-label">Progress</div>
                </div>
                <div class="status-card passed">
                    <div class="status-value passed">${suite.passed_specs}</div>
                    <div class="status-label">Passed</div>
                </div>
                <div class="status-card failed">
                    <div class="status-value failed">${suite.failed_specs}</div>
                    <div class="status-label">Failed</div>
                </div>
                <div class="status-card">
                    <div class="status-value pending">${suite.pending_specs}</div>
                    <div class="status-label">Pending</div>
                </div>
                <div class="status-card">
                    <div class="status-value">${suite.total_specs}</div>
                    <div class="status-label">Total</div>
                </div>
                <div class="status-card">
                    <div class="status-value">${Math.round(suite.running_time)}s</div>
                    <div class="status-label">Runtime</div>
                </div>
            `;

            document.getElementById('statusBar').innerHTML = html;
        }

        function renderProgress(suite) {
            const progress = suite.total_specs > 0
                ? Math.round((suite.completed_specs / suite.total_specs) * 100)
                : 0;

            const eta = suite.estimated_time_left > 0
                ? Math.round(suite.estimated_time_left) + 's'
                : 'calculating...';

            const html = `
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${progress}%">
                        ${progress}%
                    </div>
                </div>
                <div class="progress-info">
                    <span>${suite.completed_specs} / ${suite.total_specs} tests completed</span>
                    <span>ETA: ${eta}</span>
                </div>
            `;

            document.getElementById('progressSection').innerHTML = html;
        }

        function renderSpecs(specEvents) {
            const html = specEvents.slice(-20).reverse().map(event => {
                const spec = event.spec;
                const statusIcon = {
                    'running': '⏳',
                    'passed': '✅',
                    'failed': '❌',
                    'pending': '⏸️',
                    'skipped': '⏭️'
                }[spec.status] || '❓';

                const duration = spec.duration > 0 ? spec.duration.toFixed(3) + 's' : 'running...';

                let errorHtml = '';
                if (spec.error) {
                    errorHtml = `<div class="spec-error">${escapeHtml(spec.error)}</div>`;
                }

                return `
                    <div class="spec-item">
                        <div class="spec-status ${spec.status}">
                            ${statusIcon}
                        </div>
                        <div class="spec-info">
                            <div class="spec-name">${escapeHtml(spec.name)}</div>
                            <div class="spec-meta">${escapeHtml(spec.file_name)}:${spec.line_number}</div>
                            ${errorHtml}
                        </div>
                        <div class="spec-duration">${duration}</div>
                    </div>
                `;
            }).join('');

            document.getElementById('specsList').innerHTML = html || '<p>No tests running yet...</p>';
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // Load data on page load
        loadTestData();
    </script>
</body>
</html>
EOF

echo -e "${BLUE}📊 Real-time dashboard generated${NC}"
echo -e "${GREEN}📄 Dashboard: $HTML_FILE${NC}"
echo ""
echo -e "${YELLOW}💡 Usage:${NC}"
echo -e "  1. Run tests with real-time reporter"
echo -e "  2. Open $HTML_FILE in browser (auto-refreshes)"
echo -e "  3. Watch tests execute in real-time!"
echo ""
echo -e "${MAGENTA}Example:${NC}"
echo -e "  # Terminal 1: Run tests"
echo -e "  make test-unit"
echo ""
echo -e "  # Terminal 2: Watch dashboard"
echo -e "  open $HTML_FILE"
echo ""

# Check if event file exists and embed data
if [ -f "$EVENT_FILE" ]; then
    # Read event data and inject into HTML
    EVENT_DATA=$(cat "$EVENT_FILE" | sed 's/\\/\\\\/g' | sed 's/`/\\`/g')
    sed -i "s|__EVENT_DATA__|$EVENT_DATA|g" "$HTML_FILE"

    echo -e "${GREEN}✅ Dashboard updated with latest test data${NC}"
else
    echo -e "${YELLOW}⚠️  No test events yet - run tests to see live data${NC}"
fi

echo ""
echo -e "${BLUE}🌐 To view dashboard, open:${NC}"
echo -e "  file://$(pwd)/$HTML_FILE"
