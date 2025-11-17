// BusinessGlance - Chart.js Integration
// Lightweight chart rendering for business metrics

(function() {
    'use strict';

    // Simple chart rendering without external dependencies
    // Uses canvas API for lightweight metric visualizations

    window.BusinessCharts = {
        // Render a trend line chart
        renderTrendChart: function(canvasId, labels, values, options) {
            const canvas = document.getElementById(canvasId);
            if (!canvas) return;

            const ctx = canvas.getContext('2d');
            const width = canvas.width;
            const height = canvas.height;
            const padding = options?.padding || 40;

            // Clear canvas
            ctx.clearRect(0, 0, width, height);

            if (!values || values.length === 0) return;

            // Calculate scales
            const maxValue = Math.max(...values);
            const minValue = Math.min(...values);
            const range = maxValue - minValue || 1;

            const xStep = (width - 2 * padding) / (values.length - 1 || 1);
            const yScale = (height - 2 * padding) / range;

            // Draw axes
            ctx.strokeStyle = 'rgba(150, 150, 150, 0.3)';
            ctx.lineWidth = 1;

            // Y-axis
            ctx.beginPath();
            ctx.moveTo(padding, padding);
            ctx.lineTo(padding, height - padding);
            ctx.stroke();

            // X-axis
            ctx.beginPath();
            ctx.moveTo(padding, height - padding);
            ctx.lineTo(width - padding, height - padding);
            ctx.stroke();

            // Draw grid lines
            ctx.strokeStyle = 'rgba(150, 150, 150, 0.1)';
            for (let i = 0; i <= 4; i++) {
                const y = padding + (height - 2 * padding) * i / 4;
                ctx.beginPath();
                ctx.moveTo(padding, y);
                ctx.lineTo(width - padding, y);
                ctx.stroke();
            }

            // Draw line
            ctx.strokeStyle = options?.color || '#3b82f6';
            ctx.lineWidth = 2;
            ctx.beginPath();

            values.forEach((value, index) => {
                const x = padding + index * xStep;
                const y = height - padding - (value - minValue) * yScale;

                if (index === 0) {
                    ctx.moveTo(x, y);
                } else {
                    ctx.lineTo(x, y);
                }
            });

            ctx.stroke();

            // Draw points
            ctx.fillStyle = options?.color || '#3b82f6';
            values.forEach((value, index) => {
                const x = padding + index * xStep;
                const y = height - padding - (value - minValue) * yScale;

                ctx.beginPath();
                ctx.arc(x, y, 3, 0, 2 * Math.PI);
                ctx.fill();
            });

            // Draw labels
            ctx.fillStyle = 'rgba(150, 150, 150, 0.8)';
            ctx.font = '11px sans-serif';
            ctx.textAlign = 'center';

            labels.forEach((label, index) => {
                const x = padding + index * xStep;
                ctx.fillText(label, x, height - padding + 20);
            });

            // Draw value labels (top and bottom)
            ctx.textAlign = 'right';
            ctx.fillText(this.formatNumber(maxValue), padding - 5, padding + 5);
            ctx.fillText(this.formatNumber(minValue), padding - 5, height - padding + 5);
        },

        formatNumber: function(num) {
            if (num >= 1000000) {
                return (num / 1000000).toFixed(1) + 'M';
            } else if (num >= 1000) {
                return (num / 1000).toFixed(1) + 'K';
            }
            return num.toFixed(0);
        }
    };

    // Auto-render charts on page load
    document.addEventListener('DOMContentLoaded', function() {
        // Look for chart canvases and render them
        document.querySelectorAll('[data-chart-type="trend"]').forEach(function(canvas) {
            const labels = JSON.parse(canvas.dataset.labels || '[]');
            const values = JSON.parse(canvas.dataset.values || '[]');
            const color = canvas.dataset.color || '#3b82f6';

            BusinessCharts.renderTrendChart(canvas.id, labels, values, { color: color });
        });
    });
})();
