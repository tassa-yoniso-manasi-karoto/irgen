<script>
    import { onMount, onDestroy } from 'svelte';
    import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
    import ProgressBar from './ProgressBar.svelte';

    let logs = [];
    export let downloadProgress = null;
    let scrollContainer;
    let autoScroll = true;
    let isScrolling = false;
    let scrollTimeout;
    
    function getLevelClass(level) {
        const levelMap = {
            'DEBUG': 'debug',
            'INFO': 'info',
            'WARN': 'warn',
            'ERROR': 'error',
            'FATAL': 'fatal',
            'PANIC': 'panic',
            'TRACE': 'trace'
        };
        return levelMap[level] || 'info';
    }
    
    function handleScroll(e) {
        if (isScrolling) return;
        
        const target = e.currentTarget;
        const isAtBottom = Math.abs(
            target.scrollHeight - target.clientHeight - target.scrollTop
        ) < 1;
        
        // Only update autoScroll if user has manually scrolled
        if (!isScrolling) {
            autoScroll = isAtBottom;
        }

        // Clear existing timeout
        clearTimeout(scrollTimeout);
        
        // Set a new timeout
        scrollTimeout = setTimeout(() => {
            isScrolling = false;
        }, 150); // Debounce scroll events
    }

    function scrollToBottom() {
        if (!scrollContainer || !autoScroll) return;
        
        isScrolling = true;
        requestAnimationFrame(() => {
            scrollContainer.scrollTop = scrollContainer.scrollHeight;
            setTimeout(() => {
                isScrolling = false;
            }, 50);
        });
    }

    function toggleAutoScroll(value) {
        autoScroll = value;
        if (autoScroll) {
            scrollToBottom();
        }
    }

    function clearLogs() {
        logs = [];
        downloadProgress = null;
    }

    onMount(() => {
        EventsOn("log", (log) => {
            logs = [...logs, log];
            if (autoScroll) {
                scrollToBottom();
            }
        });

        EventsOn("download-progress", (progress) => {
            downloadProgress = progress;
            if (autoScroll) {
                scrollToBottom();
            }
        });
    });

    onDestroy(() => {
        EventsOff("log");
        EventsOff("download-progress");
        clearTimeout(scrollTimeout);
    });

    $: if (logs.length && autoScroll) {
        scrollToBottom();
    }
</script>

<div class="log-viewer">
    <div class="controls">
        <div class="auto-scroll">
            <button 
                type="button" 
                class="checkbox-button" 
                on:click={() => toggleAutoScroll(!autoScroll)}
                aria-pressed={autoScroll}
            >
                <input 
                    type="checkbox" 
                    checked={autoScroll}
                    on:change={(e) => toggleAutoScroll(e.target.checked)}
                    aria-hidden="true"
                />
                Auto-scroll
            </button>
        </div>
        <button on:click={clearLogs}>Clear</button>
    </div>
    
    <div class="content-wrapper">
        <div 
            class="log-container" 
            bind:this={scrollContainer} 
            on:scroll={handleScroll}
        >
            {#each logs as log}
                <div class="log-entry">
                    <span class="time">{log.time}</span>
                    <span class="level {getLevelClass(log.level)}">{log.level}</span>
                    <span class="message">{log.message}</span>
                </div>
            {/each}
        </div>
        
        {#if downloadProgress}
            <div class="progress-section">
                <ProgressBar 
                    progress={downloadProgress.progress}
                    current={downloadProgress.current}
                    total={downloadProgress.total}
                    speed={downloadProgress.speed}
                    currentFile={downloadProgress.currentFile}
                    operation={downloadProgress.operation}
                />
            </div>
        {/if}
    </div>
</div>

<style>
    .log-viewer {
        display: flex;
        flex-direction: column;
        height: 300px;
        min-height: 200px;
        max-height: 500px;
        background: #1e1e1e;
        color: #ffffff;
        font-family: monospace;
        font-size: 12px;
        border: 1px solid #333;
        border-radius: 4px;
    }

    .content-wrapper {
        display: flex;
        flex-direction: column;
        flex: 1;
        min-height: 0; /* Important for proper flexbox behavior */
        position: relative;
    }

    .controls {
        padding: 4px 8px;
        border-bottom: 1px solid #333;
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: #252525;
    }

    .auto-scroll {
        display: flex;
        align-items: center;
        gap: 4px;
    }

    .log-entry {
        padding: 2px 8px;
        border-bottom: 1px solid #2a2a2a;
        white-space: pre-wrap;
        word-wrap: break-word;
        line-height: 1.4;
        display: flex;
        align-items: baseline;
    }

    .time {
        color: #888;
        margin-right: 8px;
        font-size: 11px;
        flex-shrink: 0;
    }

    .level {
        font-weight: bold;
        margin-right: 8px;
        flex-shrink: 0;
        min-width: 40px;
    }

    .message {
        flex-grow: 1;
        color: #d4d4d4;
    }

    .progress-section {
        padding: 8px;
        background: #252525;
        border-top: 1px solid #333;
    }

    /* Log level colors */
    .debug { color: #7cafc2; }
    .info { color: #99c794; }
    .warn { color: #fac863; }
    .error { color: #ec5f67; }
    .fatal { color: #ff8080; }
    .panic { color: #ff6b6b; }
    .trace { color: #c792ea; }

    button {
        padding: 2px 8px;
        background: #333;
        border: none;
        color: #999;
        border-radius: 3px;
        cursor: pointer;
        font-size: 11px;
        text-transform: uppercase;
    }

    button:hover {
        background: #444;
        color: #fff;
    }

    input[type="checkbox"] {
        margin: 0;
    }

    label {
        font-size: 11px;
        color: #999;
    }
    
    .log-container {
        flex: 1;
        overflow-y: auto;
        padding: 0;
        min-height: 0; /* Important for proper flexbox behavior */
    }

    /* Webkit scrollbar styles - unified for both directions */
    .log-container::-webkit-scrollbar {
        width: 6px;  /* Thinner vertical scrollbar */
        height: 6px; /* Thinner horizontal scrollbar */
    }

    .log-container::-webkit-scrollbar-track {
        background: #1e1e1e;
    }

    .log-container::-webkit-scrollbar-thumb {
        background-color: #444444;
        border-radius: 3px;
        /* Remove border to make it thinner */
    }

    .log-container::-webkit-scrollbar-thumb:hover {
        background-color: #555555;
    }

    /* Style the corner where both scrollbars meet */
    .log-container::-webkit-scrollbar-corner {
        background: #1e1e1e;
    }
    
    /* Add specific styling for the checkbox to ensure it's clickable */
    .auto-scroll {
    	position: relative;
    	z-index: 10;
    }
    
    input[type="checkbox"] {
    	cursor: pointer;
    }
    
    label {
    	cursor: pointer;
    	user-select: none;
    }
    
    .progress-section {
        position: sticky;
        bottom: 0;
        left: 0;
        right: 0;
        background: #252525;
        border-top: 1px solid #333;
        z-index: 10;
    }

</style>


