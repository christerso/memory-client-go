// Dark mode toggle
const darkModeToggle = document.getElementById('dark-mode-toggle');
const body = document.body;

// Check for saved dark mode preference
if (localStorage.getItem('darkMode') === 'true') {
    body.classList.add('dark-mode');
    darkModeToggle.checked = true;
}

// Toggle dark mode
darkModeToggle.addEventListener('change', function() {
    if (darkModeToggle.checked) {
        body.classList.add('dark-mode');
        localStorage.setItem('darkMode', 'true');
    } else {
        body.classList.remove('dark-mode');
        localStorage.setItem('darkMode', 'false');
    }
});

// Memory chart
let chart;

// Load memory stats history
async function loadMemoryStats() {
    try {
        const response = await fetch('/api/memory/stats/history');
        const stats = await response.json();
        
        if (!chart) return;
        
        const chartData = stats.map(function(stat) {
            return {
                x: new Date(stat.timestamp),
                y: stat.TotalVectors
            };
        });
        
        const filesData = stats.map(function(stat) {
            return {
                x: new Date(stat.timestamp),
                y: stat.ProjectFileCount
            };
        });
        
        chart.data.datasets[0].data = chartData;
        chart.data.datasets[1].data = filesData;
        chart.update();
    } catch (error) {
        console.error('Error loading memory stats:', error);
    }
}

// Load activity log
async function loadActivityLog() {
    try {
        const response = await fetch('/api/activity/log');
        const entries = await response.json();
        
        const logContainer = document.getElementById('activity-log');
        logContainer.innerHTML = '';
        
        if (entries.length === 0) {
            logContainer.innerHTML = '<p class="text-center p-3">No activity recorded yet</p>';
            return;
        }
        
        entries.forEach(function(entry) {
            const timestamp = new Date(entry.timestamp).toLocaleTimeString();
            const logEntry = document.createElement('div');
            logEntry.className = 'log-entry px-2 py-1';
            logEntry.innerHTML = '<span class="log-timestamp">' + timestamp + '</span> ' + entry.message;
            logContainer.appendChild(logEntry);
        });
        
        // Auto-scroll to the bottom of the log
        logContainer.scrollTop = logContainer.scrollHeight;
    } catch (error) {
        console.error('Error loading activity log:', error);
    }
}

// Load project files
async function loadProjectFiles(tag) {
    try {
        let url = '/api/memory/files';
        if (tag) {
            url = '/api/memory/files/filter?tag=' + encodeURIComponent(tag);
        }
        
        const response = await fetch(url);
        const files = await response.json();
        
        const filesTable = document.getElementById('project-files');
        filesTable.innerHTML = '';
        
        document.getElementById('file-count').textContent = files.length + ' files';
        
        if (files.length === 0) {
            filesTable.innerHTML = '<tr><td colspan="4" class="text-center">No project files found</td></tr>';
            return;
        }
        
        files.forEach(function(file) {
            const row = document.createElement('tr');
            
            // Format the modified time properly
            let modTime = 'Unknown';
            if (file.mod_time) {
                modTime = new Date(file.mod_time * 1000).toLocaleString();
            }
            
            // Determine language with fallback
            const language = file.language || getLanguageFromPath(file.path);
            
            row.innerHTML = 
                '<td class="text-truncate" style="max-width: 300px;">' + file.path + '</td>' +
                '<td>' + language + '</td>' +
                '<td>' + (file.tag || '-') + '</td>' +
                '<td>' + modTime + '</td>';
            
            filesTable.appendChild(row);
        });
    } catch (error) {
        console.error('Error loading project files:', error);
    }
}

// Helper function to determine language from file path
function getLanguageFromPath(path) {
    const ext = path.split('.').pop().toLowerCase();
    const languageMap = {
        'go': 'Go',
        'js': 'JavaScript',
        'html': 'HTML',
        'css': 'CSS',
        'md': 'Markdown',
        'json': 'JSON',
        'sh': 'Shell',
        'bat': 'Batch',
        'txt': 'Text',
        'py': 'Python',
        'cpp': 'C++',
        'c': 'C',
        'h': 'C Header',
        'java': 'Java',
        'ts': 'TypeScript'
    };
    
    return languageMap[ext] || 'Text';
}

// Filter project files by tag
function filterProjectFiles() {
    const tag = document.getElementById('tag-filter').value.trim();
    if (!tag) return;
    
    const projectFiles = document.getElementById('project-files');
    const rows = projectFiles.querySelectorAll('tr');
    let visibleCount = 0;
    
    rows.forEach(row => {
        const tagCell = row.querySelector('td:nth-child(3)');
        if (!tagCell) return;
        
        const rowTag = tagCell.textContent.trim();
        if (rowTag === tag) {
            row.style.display = '';
            visibleCount++;
        } else {
            row.style.display = 'none';
        }
    });
    
    document.getElementById('file-count').textContent = `${visibleCount} files`;
}

// Clear filter and show all project files
function clearFilter() {
    const projectFiles = document.getElementById('project-files');
    const rows = projectFiles.querySelectorAll('tr');
    let totalCount = 0;
    
    rows.forEach(row => {
        row.style.display = '';
        totalCount++;
    });
    
    document.getElementById('tag-filter').value = '';
    document.getElementById('file-count').textContent = `${totalCount} files`;
}

// Clear memory
async function clearMemory(type) {
    if (!confirm('Are you sure you want to clear ' + type + ' memories?')) {
        return;
    }
    
    try {
        let url = '';
        switch (type) {
            case 'all':
                url = '/api/memory/clear/all';
                break;
            case 'messages':
                url = '/api/memory/clear/messages';
                break;
            case 'project-files':
                url = '/api/memory/clear/files';
                break;
            default:
                console.error('Unknown memory type:', type);
                return;
        }
        
        const response = await fetch(url, { method: 'POST' });
        if (response.ok) {
            alert(type + ' memories cleared successfully');
            // Reload data
            loadMemoryStats();
            loadActivityLog();
            loadProjectFiles();
        } else {
            alert('Failed to clear ' + type + ' memories');
        }
    } catch (error) {
        console.error('Error clearing memories:', error);
        alert('Error clearing memories: ' + error.message);
    }
}

// Update uptime
function updateUptime() {
    const uptimeElement = document.getElementById('uptime');
    if (!uptimeElement) return;
    
    fetch('/api/uptime')
        .then(response => response.text())
        .then(uptime => {
            uptimeElement.textContent = uptime;
        })
        .catch(error => {
            console.error('Error fetching uptime:', error);
        });
}

// Initialize
document.addEventListener('DOMContentLoaded', function() {
    // Get chart instance from the global scope
    chart = window.chart;
    
    // Load initial data
    loadMemoryStats();
    loadActivityLog();
    loadProjectFiles();
    
    // Set up refresh intervals
    // Refresh memory stats every 5 seconds for a smoother chart
    setInterval(loadMemoryStats, 5000);
    
    // Refresh activity log every 10 seconds
    setInterval(loadActivityLog, 10000);
    
    // Refresh project files every 30 seconds
    setInterval(loadProjectFiles, 30000);
    
    // Update uptime every second
    setInterval(updateUptime, 1000);
    
    // Set up event listeners
    document.getElementById('filter-button').addEventListener('click', filterProjectFiles);
    document.getElementById('clear-filter-button').addEventListener('click', clearFilter);
    document.getElementById('clear-all-button').addEventListener('click', function() { clearMemory('all'); });
    document.getElementById('clear-messages-button').addEventListener('click', function() { clearMemory('messages'); });
    document.getElementById('clear-files-button').addEventListener('click', function() { clearMemory('project-files'); });
});
