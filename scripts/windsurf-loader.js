// Windsurf Memory Integration Loader
// This script loads the memory integration into Windsurf

(function() {
    // Configuration
    const config = {
        // Choose which integration to load
        useStandalone: true, // Set to true to use standalone version, false to use API version
        
        // Script paths - update these if needed
        standaloneScriptPath: 'file:///C:/Program Files/MemoryClientMCP/windsurf-standalone-memory.js',
        apiScriptPath: 'file:///C:/Program Files/MemoryClientMCP/windsurf-memory-integration.js'
    };

    // Load the appropriate script
    function loadMemoryIntegration() {
        const scriptPath = config.useStandalone ? config.standaloneScriptPath : config.apiScriptPath;
        
        console.log(`Loading Windsurf Memory Integration from: ${scriptPath}`);
        
        const script = document.createElement('script');
        script.src = scriptPath;
        script.onerror = () => {
            console.error(`Failed to load memory integration from ${scriptPath}`);
            // Show error notification if possible
            if (typeof windsurf !== 'undefined' && windsurf.notifications) {
                windsurf.notifications.show({
                    message: `Failed to load memory integration from ${scriptPath}`,
                    type: 'error',
                    duration: 5000
                });
            }
        };
        script.onload = () => {
            console.log('Memory integration loaded successfully');
            // Show success notification if possible
            if (typeof windsurf !== 'undefined' && windsurf.notifications) {
                windsurf.notifications.show({
                    message: 'Memory integration loaded successfully',
                    type: 'success',
                    duration: 3000
                });
            }
        };
        
        document.head.appendChild(script);
    }

    // Initialize when the document is ready
    if (document.readyState === 'complete') {
        loadMemoryIntegration();
    } else {
        document.addEventListener('DOMContentLoaded', loadMemoryIntegration);
    }
})();
