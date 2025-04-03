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
async function loadProjectFiles() {
    try {
        const response = await fetch('/api/project-files');
        const data = await response.json();
        
        const tableBody = document.getElementById('projectFilesTable');
        tableBody.innerHTML = '';
        
        if (data.files && data.files.length > 0) {
            data.files.forEach(function(file) {
                const row = document.createElement('tr');
                
                // Extract filename from path
                const filename = file.path.split('/').pop();
                
                // Format file size
                const size = formatFileSize(file.size);
                
                // Format date
                const date = new Date(file.modified);
                const formattedDate = date.toLocaleString();
                
                row.innerHTML = `
                    <td>${filename}</td>
                    <td>${file.path}</td>
                    <td>${size}</td>
                    <td>${formattedDate}</td>
                `;
                
                tableBody.appendChild(row);
            });
        } else {
            const row = document.createElement('tr');
            row.innerHTML = '<td colspan="4" class="text-center">No project files found</td>';
            tableBody.appendChild(row);
        }
    } catch (error) {
        console.error('Error loading project files:', error);
        const tableBody = document.getElementById('projectFilesTable');
        tableBody.innerHTML = '<tr><td colspan="4" class="text-center">Error loading project files</td></tr>';
    }
}

// Format file size
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Load conversation history
async function loadConversationHistory() {
    try {
        const response = await fetch('/api/conversation-history?limit=50');
        const data = await response.json();
        
        const tableBody = document.getElementById('conversationHistoryTable');
        tableBody.innerHTML = '';
        
        if (data.messages && data.messages.length > 0) {
            // Sort messages by timestamp (newest first)
            data.messages.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
            
            data.messages.forEach(function(message) {
                const row = document.createElement('tr');
                
                // Format date
                const date = new Date(message.timestamp);
                const formattedDate = date.toLocaleString();
                
                // Format role with color
                let roleClass = '';
                switch(message.role) {
                    case 'user':
                        roleClass = 'text-primary';
                        break;
                    case 'assistant':
                        roleClass = 'text-success';
                        break;
                    case 'system':
                        roleClass = 'text-warning';
                        break;
                    default:
                        roleClass = 'text-secondary';
                }
                
                row.innerHTML = `
                    <td>${formattedDate}</td>
                    <td><span class="${roleClass}">${message.role}</span></td>
                    <td>${escapeHtml(message.content)}</td>
                `;
                
                tableBody.appendChild(row);
            });
        } else {
            const row = document.createElement('tr');
            row.innerHTML = '<td colspan="3" class="text-center">No messages found</td>';
            tableBody.appendChild(row);
        }
    } catch (error) {
        console.error('Error loading conversation history:', error);
        const tableBody = document.getElementById('conversationHistoryTable');
        tableBody.innerHTML = '<tr><td colspan="3" class="text-center">Error loading conversation history</td></tr>';
    }
}

// Escape HTML to prevent XSS
function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
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
    loadConversationHistory();
    
    // Set up refresh buttons
    document.querySelector('.refresh-files-btn').addEventListener('click', loadProjectFiles);
    document.querySelector('.refresh-history-btn').addEventListener('click', loadConversationHistory);
    
    // Set up auto-refresh
    setInterval(loadMemoryStats, 15000);
    setInterval(loadActivityLog, 15000);
    setInterval(loadConversationHistory, 15000);
    setInterval(updateUptime, 1000);
    
    // Set up event listeners for memory clearing
    document.getElementById('clear-vectors-btn').addEventListener('click', function() {
        clearMemory('vectors');
    });
    
    document.getElementById('clear-files-btn').addEventListener('click', function() {
        clearMemory('files');
    });
    
    document.getElementById('clear-all-btn').addEventListener('click', function() {
        clearMemory('all');
    });
});
