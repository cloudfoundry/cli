#!/bin/bash
# Test Dependency Visualizer
# Creates interactive visualization of test dependencies

set -e

REPORT_DIR="test-reports/dependencies"
OUTPUT_FILE="$REPORT_DIR/dependency-graph.html"
DOT_FILE="$REPORT_DIR/dependencies.dot"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

echo "=========================================="
echo "🕸️  Test Dependency Visualizer"
echo "=========================================="
echo ""

echo "🔍 Analyzing test dependencies..."

# Analyze Go module dependencies
echo "digraph TestDependencies {" > "$DOT_FILE"
echo "  rankdir=LR;" >> "$DOT_FILE"
echo "  node [shape=box, style=filled, fillcolor=lightblue];" >> "$DOT_FILE"
echo "" >> "$DOT_FILE"

# Find all test files and their imports
declare -A test_packages
declare -A dependencies

total_packages=0
total_dependencies=0

while IFS= read -r test_file; do
    package_name=$(dirname "$test_file" | sed 's|^\./||')

    if [ -z "${test_packages[$package_name]}" ]; then
        test_packages[$package_name]=1
        total_packages=$((total_packages + 1))
    fi

    # Extract imports from test file
    while IFS= read -r import_line; do
        if [[ $import_line =~ \"(.+)\" ]]; then
            imported_pkg="${BASH_REMATCH[1]}"

            # Only include local imports (containing github.com/cloudfoundry/cli)
            if [[ $imported_pkg == *"github.com/cloudfoundry/cli"* ]]; then
                # Shorten package name
                short_pkg=$(echo "$imported_pkg" | sed 's|github.com/cloudfoundry/cli/||')

                dep_key="$package_name -> $short_pkg"
                if [ -z "${dependencies[$dep_key]}" ]; then
                    dependencies[$dep_key]=1
                    total_dependencies=$((total_dependencies + 1))

                    # Add edge to DOT file
                    echo "  \"$package_name\" -> \"$short_pkg\";" >> "$DOT_FILE"
                fi
            fi
        fi
    done < <(grep -E "^\s*(import|\")" "$test_file" 2>/dev/null | grep -v "^import (")
done < <(find . -name "*_test.go" -type f 2>/dev/null | head -100)

echo "}" >> "$DOT_FILE"

echo ""
echo "=========================================="
echo "📊 Dependency Statistics"
echo "=========================================="
echo -e "Test Packages:     $total_packages"
echo -e "Dependencies:      $total_dependencies"
echo ""

# Generate interactive HTML visualization
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Test Dependency Graph</title>
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
        }

        .header {
            background: linear-gradient(135deg, #bd34fe 0%, #41d1ff 100%);
            padding: 30px;
            text-align: center;
            color: white;
        }

        .header h1 {
            font-size: 36px;
            margin-bottom: 10px;
        }

        .container {
            padding: 20px;
            max-width: 1600px;
            margin: 0 auto;
        }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }

        .stat-card {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
        }

        .stat-value {
            font-size: 48px;
            font-weight: bold;
            color: #58a6ff;
        }

        .stat-label {
            color: #8b949e;
            font-size: 14px;
            margin-top: 5px;
        }

        #graph-container {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 20px;
            min-height: 600px;
            position: relative;
        }

        .node {
            cursor: pointer;
            transition: all 0.3s;
        }

        .node:hover {
            opacity: 0.8;
        }

        .node-circle {
            fill: #58a6ff;
            stroke: #1f6feb;
            stroke-width: 2px;
        }

        .node-text {
            fill: #c9d1d9;
            font-size: 12px;
            font-family: 'Courier New', monospace;
            pointer-events: none;
        }

        .link {
            stroke: #30363d;
            stroke-width: 1.5px;
            fill: none;
            marker-end: url(#arrowhead);
        }

        .link:hover {
            stroke: #58a6ff;
            stroke-width: 2px;
        }

        .controls {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 20px;
        }

        .controls button {
            background: #238636;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            margin-right: 10px;
        }

        .controls button:hover {
            background: #2ea043;
        }

        .legend {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 15px;
            margin-top: 20px;
        }

        .legend h3 {
            color: #58a6ff;
            margin-bottom: 10px;
        }

        .legend-item {
            display: flex;
            align-items: center;
            margin: 8px 0;
            color: #8b949e;
        }

        .legend-color {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            margin-right: 10px;
        }
    </style>
    <script src="https://d3js.org/d3.v7.min.js"></script>
</head>
<body>
    <div class="header">
        <h1>🕸️ Test Dependency Graph</h1>
        <p>Interactive Visualization of Test Package Dependencies</p>
    </div>

    <div class="container">
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">EOF

echo "$total_packages" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<EOF
</div>
                <div class="stat-label">Test Packages</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">$total_dependencies</div>
                <div class="stat-label">Dependencies</div>
            </div>
        </div>

        <div class="controls">
            <button onclick="resetZoom()">Reset Zoom</button>
            <button onclick="centerGraph()">Center</button>
            <button onclick="exportSVG()">Export SVG</button>
        </div>

        <div id="graph-container"></div>

        <div class="legend">
            <h3>Legend</h3>
            <div class="legend-item">
                <div class="legend-color" style="background: #58a6ff;"></div>
                <span>Test Package</span>
            </div>
            <div class="legend-item">
                <div class="legend-color" style="background: #30363d; border-radius: 0;"></div>
                <span>Dependency</span>
            </div>
        </div>
    </div>

    <script>
        // Parse dependencies
        const dependencies = \`__DEPENDENCIES__\`.trim().split('\\n')
            .filter(line => line.includes('->'))
            .map(line => {
                const match = line.match(/"(.+)"\s*->\s*"(.+)"/);
                if (match) {
                    return { source: match[1], target: match[2] };
                }
                return null;
            })
            .filter(d => d !== null);

        // Build graph data
        const nodes = new Map();
        dependencies.forEach(d => {
            if (!nodes.has(d.source)) {
                nodes.set(d.source, { id: d.source, group: 1 });
            }
            if (!nodes.has(d.target)) {
                nodes.set(d.target, { id: d.target, group: 2 });
            }
        });

        const graphData = {
            nodes: Array.from(nodes.values()),
            links: dependencies
        };

        // Create SVG
        const width = document.getElementById('graph-container').clientWidth;
        const height = 600;

        const svg = d3.select('#graph-container')
            .append('svg')
            .attr('width', width)
            .attr('height', height)
            .call(d3.zoom().on('zoom', (event) => {
                container.attr('transform', event.transform);
            }));

        // Add arrow marker
        svg.append('defs').append('marker')
            .attr('id', 'arrowhead')
            .attr('viewBox', '-0 -5 10 10')
            .attr('refX', 20)
            .attr('refY', 0)
            .attr('orient', 'auto')
            .attr('markerWidth', 8)
            .attr('markerHeight', 8)
            .append('svg:path')
            .attr('d', 'M 0,-5 L 10 ,0 L 0,5')
            .attr('fill', '#30363d');

        const container = svg.append('g');

        // Create force simulation
        const simulation = d3.forceSimulation(graphData.nodes)
            .force('link', d3.forceLink(graphData.links).id(d => d.id).distance(100))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2))
            .force('collision', d3.forceCollide().radius(30));

        // Draw links
        const link = container.append('g')
            .selectAll('path')
            .data(graphData.links)
            .join('path')
            .attr('class', 'link');

        // Draw nodes
        const node = container.append('g')
            .selectAll('g')
            .data(graphData.nodes)
            .join('g')
            .attr('class', 'node')
            .call(d3.drag()
                .on('start', dragstarted)
                .on('drag', dragged)
                .on('end', dragended));

        node.append('circle')
            .attr('class', 'node-circle')
            .attr('r', 8);

        node.append('text')
            .attr('class', 'node-text')
            .attr('dx', 12)
            .attr('dy', 4)
            .text(d => d.id.split('/').pop());

        // Update positions
        simulation.on('tick', () => {
            link.attr('d', d => {
                const dx = d.target.x - d.source.x;
                const dy = d.target.y - d.source.y;
                const dr = Math.sqrt(dx * dx + dy * dy);
                return \`M\${d.source.x},\${d.source.y}L\${d.target.x},\${d.target.y}\`;
            });

            node.attr('transform', d => \`translate(\${d.x},\${d.y})\`);
        });

        // Drag functions
        function dragstarted(event) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            event.subject.fx = event.subject.x;
            event.subject.fy = event.subject.y;
        }

        function dragged(event) {
            event.subject.fx = event.x;
            event.subject.fy = event.y;
        }

        function dragended(event) {
            if (!event.active) simulation.alphaTarget(0);
            event.subject.fx = null;
            event.subject.fy = null;
        }

        // Control functions
        function resetZoom() {
            svg.transition().duration(750).call(
                d3.zoom().transform,
                d3.zoomIdentity
            );
        }

        function centerGraph() {
            const bounds = container.node().getBBox();
            const fullWidth = width;
            const fullHeight = height;
            const midX = bounds.x + bounds.width / 2;
            const midY = bounds.y + bounds.height / 2;

            svg.transition().duration(750).call(
                d3.zoom().transform,
                d3.zoomIdentity
                    .translate(fullWidth / 2, fullHeight / 2)
                    .scale(0.8)
                    .translate(-midX, -midY)
            );
        }

        function exportSVG() {
            const svgData = document.querySelector('#graph-container svg').outerHTML;
            const blob = new Blob([svgData], { type: 'image/svg+xml' });
            const url = URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = 'test-dependencies.svg';
            link.click();
        }
    </script>
</body>
</html>
EOF

# Embed DOT file content
if [ -f "$DOT_FILE" ]; then
    DOT_CONTENT=$(cat "$DOT_FILE" | sed 's/\\/\\\\/g' | sed 's/`/\\`/g')
    sed -i "s|__DEPENDENCIES__|$DOT_CONTENT|g" "$OUTPUT_FILE"
fi

echo -e "${GREEN}✅ Dependency visualization complete!${NC}"
echo -e "${BLUE}📄 HTML Report: $OUTPUT_FILE${NC}"
echo -e "${BLUE}📄 DOT File: $DOT_FILE${NC}"
echo ""
echo -e "${MAGENTA}💡 Usage:${NC}"
echo -e "  - Open HTML for interactive graph"
echo -e "  - Use DOT file with Graphviz tools"
echo ""
echo -e "${BLUE}🌐 View visualization:${NC}"
echo -e "  open $OUTPUT_FILE"
