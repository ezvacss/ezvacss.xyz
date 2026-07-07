document.addEventListener('DOMContentLoaded', () => {
    
    const refreshBtn = document.getElementById('refresh-stats-btn');
    const refreshIcon = document.getElementById('refresh-icon');
    const statLinks = document.getElementById('stat-value-links');
    const statClicks = document.getElementById('stat-value-clicks');
    const statAvg = document.getElementById('stat-value-avg');
    const statLogs = document.getElementById('stat-value-logs');
    const logsTableBody = document.getElementById('logs-table-body');
    const searchInput = document.getElementById('log-search-input');

    
    let clicksChart = null;
    let deviceChart = null;
    let allLogs = []; 
    let filteredLogs = []; 
    let currentPage = 1;
    const rowsPerPage = 10;

    
    async function loadMetrics() {

        try {
            
            const userRes = await fetch('/api/user');
            if (!userRes.ok) {
                window.location.href = '/login';
                return;
            }
            const user = await userRes.json();
            if (!user.authenticated) {
                window.location.href = '/login';
                return;
            }
            
            
            const userDisplay = document.getElementById('user-display');
            const logoutBtn = document.getElementById('logout-btn');
            if (userDisplay) {
                userDisplay.textContent = `@${user.username}`;
                userDisplay.style.display = 'inline-block';
            }
            if (logoutBtn) {
                logoutBtn.style.display = 'inline-block';
            }

            const navMenuBar = document.getElementById('nav-menu-bar');
            if (navMenuBar) {
                navMenuBar.style.opacity = '1';
            }

            const response = await fetch('/api/stats');
            if (!response.ok) {
                throw new Error('Failed to fetch analytics data');
            }

            const data = await response.json();
            allLogs = data.recent_logs || [];
            filteredLogs = allLogs;
            currentPage = 1;

            
            statLinks.textContent = data.total_links || 0;
            statClicks.textContent = data.total_clicks || 0;
            
            const avg = data.total_links > 0 ? (data.total_clicks / data.total_links) : 0;
            statAvg.textContent = avg.toFixed(1);
            statLogs.textContent = allLogs.length;

            
            renderCharts(data.links || []);

            
            renderLogsTable(filteredLogs);

        } catch (error) {
            console.error('Error fetching analytics:', error);
            logsTableBody.innerHTML = `
                <tr>
                    <td colspan="6" style="text-align: center; color: var(--color-danger); padding: 3rem;">
                        <i data-lucide="alert-triangle" style="margin-bottom: 0.5rem; display: block; margin-left: auto; margin-right: auto; width: 2rem; height: 2rem;"></i>
                        Failed to connect to analytics API. Is the server running?
                    </td>
                </tr>
            `;
            lucide.createIcons();
        } finally {
            
        }
    }

    
    function getBrowserInfo(ua) {
        if (!ua) return 'Unknown / Script';
        const uaLower = ua.toLowerCase();
        
        
        let browser = 'Other Browser';
        let platform = 'Desktop';

        if (uaLower.includes('firefox')) browser = 'Firefox';
        else if (uaLower.includes('chrome')) browser = 'Chrome';
        else if (uaLower.includes('safari')) browser = 'Safari';
        else if (uaLower.includes('edge') || uaLower.includes('edg')) browser = 'Edge';
        else if (uaLower.includes('opr') || uaLower.includes('opera')) browser = 'Opera';

        if (uaLower.includes('mobile') || uaLower.includes('android') || uaLower.includes('iphone') || uaLower.includes('ipad')) {
            platform = 'Mobile';
        }

        return { browser, platform };
    }

    
    function renderCharts(linksData) {
        
        if (clicksChart) clicksChart.destroy();
        if (deviceChart) deviceChart.destroy();

        
        const topLinks = linksData.slice(0, 10); 
        const barLabels = topLinks.map(l => `/${l.short_code}`);
        const barData = topLinks.map(l => l.clicks);

        const ctxBar = document.getElementById('clicksChart').getContext('2d');
        clicksChart = new Chart(ctxBar, {
            type: 'bar',
            data: {
                labels: barLabels.length ? barLabels : ['No Data'],
                datasets: [{
                    label: 'Clicks',
                    data: barData.length ? barData : [0],
                    backgroundColor: '#a1a1aa',
                    borderColor: '#a1a1aa',
                    borderWidth: 1.5,
                    borderRadius: 0,
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: { color: '#3f3f46' },
                        ticks: { color: '#a1a1aa' }
                    },
                    x: {
                        grid: { display: false },
                        ticks: { color: '#a1a1aa' }
                    }
                }
            }
        });

        
        let mobileCount = 0;
        let desktopCount = 0;

        allLogs.forEach(log => {
            if (log.action === 'redirect') {
                const info = getBrowserInfo(log.user_agent);
                if (info.platform === 'Mobile') mobileCount++;
                else desktopCount++;
            }
        });

        
        if (mobileCount === 0 && desktopCount === 0) {
            desktopCount = 1; 
        }

        const ctxDoughnut = document.getElementById('deviceChart').getContext('2d');
        deviceChart = new Chart(ctxDoughnut, {
            type: 'doughnut',
            data: {
                labels: ['Desktop', 'Mobile'],
                datasets: [{
                    data: [desktopCount, mobileCount],
                    backgroundColor: ['#e4e4e7', '#3f3f46'],
                    borderColor: ['#27272a', '#27272a'],
                    borderWidth: 1.5,
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: { color: '#e4e4e7', boxWidth: 12 }
                    }
                },
                cutout: '70%'
            }
        });
    }

    

    
    function escapeHTML(str) {
        if (!str) return '';
        return str.replace(/[&<>"']/g, function(m) {
            switch (m) {
                case '&': return '&amp;';
                case '<': return '&lt;';
                case '>': return '&gt;';
                case '"': return '&quot;';
                case "'": return '&#039;';
                default: return m;
            }
        });
    }

    function renderLogsTable(logsList) {
        const prevBtn = document.getElementById('prev-page-btn');
        const nextBtn = document.getElementById('next-page-btn');
        const pageInfo = document.getElementById('page-info');

        if (!logsList || logsList.length === 0) {
            logsTableBody.innerHTML = `
                <tr>
                    <td colspan="6" style="text-align: center; color: var(--text-muted); padding: 3rem;">
                        <i data-lucide="database-backup" style="margin-bottom: 0.5rem; display: block; margin-left: auto; margin-right: auto; width: 2rem; height: 2rem;"></i>
                        No database activity logged yet. Shorten some links!
                    </td>
                </tr>
            `;
            if (prevBtn) prevBtn.disabled = true;
            if (nextBtn) nextBtn.disabled = true;
            if (pageInfo) pageInfo.textContent = 'Page 1 of 1';
            lucide.createIcons();
            return;
        }

        const totalPages = Math.ceil(logsList.length / rowsPerPage);
        if (currentPage > totalPages) currentPage = totalPages;
        if (currentPage < 1) currentPage = 1;

        const startIndex = (currentPage - 1) * rowsPerPage;
        const endIndex = Math.min(startIndex + rowsPerPage, logsList.length);
        const paginatedLogs = logsList.slice(startIndex, endIndex);

        logsTableBody.innerHTML = paginatedLogs.map(log => {
            const formattedTime = new Date(log.created_at).toLocaleString();
            
            
            let actionBadgeStyle = 'color: var(--color-primary); background: rgba(168, 85, 247, 0.1); border: 1px solid rgba(168, 85, 247, 0.2);';
            if (log.action === 'redirect') {
                actionBadgeStyle = 'color: var(--color-success); background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.2);';
            }

            
            const uaInfo = getBrowserInfo(log.user_agent);
            const userAgentString = typeof uaInfo === 'object' ? `${uaInfo.browser} (${uaInfo.platform})` : uaInfo;

            const locationText = log.location ? log.location : 'Resolving...';

            const userAgentSafe = escapeHTML(log.user_agent || '');
            const userAgentStringSafe = escapeHTML(userAgentString);
            const locationSafe = escapeHTML(locationText);
            const shortCodeSafe = escapeHTML(log.short_code);

            return `
                <tr>
                    <td>#${log.id}</td>
                    <td style="white-space: nowrap;">${formattedTime}</td>
                    <td><a href="/${shortCodeSafe}" target="_blank" class="code-badge" style="color: var(--text-primary); text-decoration: underline; font-weight: 700;">/${shortCodeSafe}</a></td>
                    <td>
                        <span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 6px; font-size: 0.8rem; font-weight: 600; text-transform: uppercase; ${actionBadgeStyle}">
                            ${log.action}
                        </span>
                    </td>
                    <td class="location-cell" style="font-family: monospace; color: var(--text-secondary);" title="${locationSafe}">${locationSafe}</td>
                    <td style="max-width: 250px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${userAgentSafe}">
                        ${userAgentStringSafe}
                    </td>
                </tr>
            `;
        }).join('');

        
        if (prevBtn) prevBtn.disabled = currentPage === 1;
        if (nextBtn) nextBtn.disabled = currentPage === totalPages || totalPages === 0;
        if (pageInfo) pageInfo.textContent = `Page ${currentPage} of ${totalPages || 1}`;

        lucide.createIcons();
    }

    
    searchInput.addEventListener('input', (e) => {
        const query = e.target.value.toLowerCase().trim();
        if (!query) {
            filteredLogs = allLogs;
            currentPage = 1;
            renderLogsTable(filteredLogs);
            return;
        }

        const filtered = allLogs.filter(log => {
            const codeMatches = log.short_code.toLowerCase().includes(query);
            const actionMatches = log.action.toLowerCase().includes(query);
            const ipMatches = (log.ip_address || '').toLowerCase().includes(query);
            const uaMatches = (log.user_agent || '').toLowerCase().includes(query);
            
            return codeMatches || actionMatches || ipMatches || uaMatches;
        });

        filteredLogs = filtered;
        currentPage = 1;
        renderLogsTable(filteredLogs);
    });

    
    const prevPageBtn = document.getElementById('prev-page-btn');
    if (prevPageBtn) {
        prevPageBtn.addEventListener('click', () => {
            if (currentPage > 1) {
                currentPage--;
                renderLogsTable(filteredLogs);
            }
        });
    }

    const nextPageBtn = document.getElementById('next-page-btn');
    if (nextPageBtn) {
        nextPageBtn.addEventListener('click', () => {
            const totalPages = Math.ceil(filteredLogs.length / rowsPerPage);
            if (currentPage < totalPages) {
                currentPage++;
                renderLogsTable(filteredLogs);
            }
        });
    }

    
    loadMetrics();

    
    if (refreshBtn) {
        refreshBtn.addEventListener('click', () => {
            window.location.reload();
        });
    }

    
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            try {
                const res = await fetch('/api/logout', { method: 'POST' });
                if (res.ok) {
                    window.location.href = '/';
                }
            } catch (err) {
                console.error('Logout failed:', err);
            }
        });
    }
});
