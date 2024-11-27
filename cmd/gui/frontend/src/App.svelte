<script lang="ts">
    import { onMount } from 'svelte';
    import "./app.css"
    import LogViewer from './LogViewer.svelte';
    import ThemeToggle from './ThemeToggle.svelte';
    import Alert from './Alert.svelte';
    
    let url = '';
    let numberOfTitle = 3;
    let maxXResolution = 1920;
    let maxYResolution = 1080;
    let status = '';
    let isProcessing = false;
    let isDark = true;
    let version = "0.0.0-n/a";
    let downloadProgress = null;
    let ankiConnectError: string | null = null;
    let alertVisible = false;
    
    onMount(async () => {
        version = await window.go.gui.App.GetVersion();
        await checkAnkiConnect();
    });
    
    async function checkAnkiConnect() {
        try {
            const result = await window.go.gui.App.QueryAnkiConnect4MediaDir({
                action: "getMediaDirPath"
            });
            if (!result) {
                ankiConnectError = "Failed to connect to AnkiConnect. Please make sure Anki is running and AnkiConnect is properly installed.";
                alertVisible = true;
            }
        } catch (error) {
            ankiConnectError = error.message || "An error occurred while connecting to AnkiConnect";
            alertVisible = true;
        }
    }
    
    async function openFileDialog() {
        try {
            const filepath = await window.go.gui.App.OpenFileDialog();
            if (filepath) {
                url = filepath;
            }
        } catch (error) {
            console.error('Error opening file dialog:', error);
        }
    }
    
    async function processURL() {
        if (!url) {
            alert('Please enter a URL');
            return;
        }

        if (numberOfTitle < 1) {
            alert('Number of Title must be at least 1');
            return;
        }

        if (maxXResolution < 1 || maxYResolution < 1) {
            alert('Resolution values must be positive');
            return;
        }

        isProcessing = true;
	downloadProgress = null;
	
        try {
            const params = {
                url: url,
                numberOfTitle: parseInt(numberOfTitle),
                maxXResolution: parseInt(maxXResolution),
                maxYResolution: parseInt(maxYResolution)
            };
            
            status = await window.go.gui.App.Process(params);
        } catch (error) {
            console.error('Error processing URL:', error);
            status = 'Error processing URL';
        } finally {
            isProcessing = false;
        }
    }
    
    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === 'Enter' && !isProcessing) {
            event.preventDefault(); // Prevent default form submission
            processURL();
        }
    }
</script>

<div class="theme-toggle-wrapper">
    <span class="version">irgen {version}</span>
    <ThemeToggle bind:isDark />
</div>

<div class="app-background" class:dark={isDark}>
    <main class="container">
        {#if ankiConnectError}
            <Alert 
                message={ankiConnectError}
                type="error"
                bind:visible={alertVisible}
                on:dismiss={() => alertVisible = false}
            />
        {/if}
	<div class="url-input-container">
	<div class="url-input">
	    <input
		type="text"
		bind:value={url}
		placeholder="Enter URL or path to HTML file here"
		on:keydown={handleKeyDown}
	    />
	    <button 
		class="file-picker-btn" 
		on:click={openFileDialog}
		title="Choose HTML file"
	    >
		    <svg 
		        xmlns="http://www.w3.org/2000/svg" 
		        width="20" 
		        height="20" 
		        viewBox="0 0 24 24" 
		        fill="none" 
		        stroke="currentColor" 
		        stroke-width="2" 
		        stroke-linecap="round" 
		        stroke-linejoin="round"
		    >
		        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
		        <polyline points="14 2 14 8 20 8"></polyline>
		        <line x1="12" y1="18" x2="12" y2="12"></line>
		        <line x1="9" y1="15" x2="15" y2="15"></line>
		    </svg>
		</button>
	    </div>
	</div>

	<div class="controls-row">
	    <div class="number-inputs">
		<div class="input-group">
		    <label for="numberOfTitle">Number of Title</label>
		    <input
		        type="number"
		        id="numberOfTitle"
		        bind:value={numberOfTitle}
		        min="1"
		    />
		</div>

		<div class="input-group">
		    <label for="maxXResolution">Max X Resolution</label>
		    <input
		        type="number"
		        id="maxXResolution"
		        bind:value={maxXResolution}
		        min="1"
		    />
		</div>

		<div class="input-group">
		    <label for="maxYResolution">Max Y Resolution</label>
		    <input
		        type="number"
		        id="maxYResolution"
		        bind:value={maxYResolution}
		        min="1"
		    />
		</div>
	    </div>

	    <div class="input-group process-group">
		<label> </label>
		<button 
		    class="process-button" 
		    on:click={processURL} 
		    disabled={isProcessing}
		    title={isProcessing ? "Processing..." : "Start processing"}
		>
		    {#if isProcessing}
		        Processing
		        <div class="spinner">&nbsp;</div>
		    {:else}
		        Process
		    {/if}
		</button>
	    </div>
	</div>
        
        <LogViewer bind:downloadProgress />

        {#if status}
            <div class="status">
                {status}
            </div>
        {/if}
    </main>
</div>

<style>
    :global(body) {
        margin: 0;
        padding: 0;
        min-height: 100vh;
    }

    :global(:root) {
        /* Light theme variables */
        --bg-color: #cce0f5;
        --container-bg: #ffffff;
        --text-color: #333333;
        --input-border: #cccccc;
        --input-bg: #ffffff;
        --button-bg: #56c865;
        --button-hover: #4CAF50;
        --hover-color: rgba(0, 0, 0, 0.1);
    }

    :global(.dark) {
        /* Dark theme variables */
        --bg-color: #1a1a1a;
        --container-bg: #2d2d2d;
        --text-color: #e0e0e0;
        --input-border: #404040;
        --input-bg: #333333;
        /* Warmer button colors for dark mode */
        --button-bg: #ed924f;
        --button-hover: #ed6500;
        --hover-color: rgba(255, 255, 255, 0.1);
    }

    /* Theme toggle wrapper styles */
    .theme-toggle-wrapper {
        position: fixed;
        top: 0;
        right: 0;
        z-index: 1000;
        padding: 1rem;
        background: transparent;
    }

    .app-background {
        min-height: 100vh;
        width: 100%;
        background-color: var(--bg-color);
        box-sizing: border-box;
        transition: background-color 0.3s;
        padding-top: 0.7rem; /* Add padding to prevent content from being hidden under the toggle */
    }

    .container {
        max-width: 1200px;
        margin: 0 auto;
        padding: 1rem;
    }

    .url-input-container {
        margin-bottom: 1rem;
        width: 100%;
    }

    .url-input {
        display: flex;
        gap: 0.5rem;
        width: 100%;
    }

    .url-input input {
        flex: 1;
        padding: 0.75rem;
        border: 1px solid var(--input-border);
        border-radius: 4px;
        font-size: 1rem;
        background-color: var(--input-bg);
        color: var(--text-color);
    }

    .file-picker-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.5rem;
        background-color: var(--input-bg);
        border: 1px solid var(--input-border);
        border-radius: 4px;
        color: var(--text-color);
        cursor: pointer;
        transition: all 0.2s ease-in-out;
        min-width: 42px;
    }

    .file-picker-btn:hover {
        background-color: var(--button-bg);
        border-color: var(--button-bg);
        color: white;
    }

    .file-picker-btn:active {
        transform: scale(0.95);
    }

    .url-input input::placeholder {
        color: var(--text-color);
        opacity: 0.6;
    }

    .controls-row {
        display: flex;
        align-items: flex-start;
        gap: 1.5rem;
        margin-bottom: 1rem;
        width: 100%;
    }

    .number-inputs {
        display: flex;
        gap: 1rem;
        flex: 1;
        max-width: calc(100% - 200px); /* Reserve space for process button */
    }

    .input-group {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        min-width: 0; /* Allow flex items to shrink below content size */
    }

    label {
        font-size: 0.75rem;
        font-weight: 500;
        color: var(--text-color);
        white-space: nowrap;
        opacity: 0.9;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    input[type="number"] {
        width: 100%;
        padding: 0.375rem 0.5rem;
        border: 1px solid var(--input-border);
        border-radius: 4px;
        text-align: center;
        font-size: 0.875rem;
        background-color: var(--input-bg);
        color: var(--text-color);
        transition: all 0.2s ease;
    }
    
    input[type="number"]:focus {
        border-color: var(--button-bg);
        outline: none;
        box-shadow: 0 0 0 2px rgba(237, 146, 79, 0.2);
    }

    .process-button {
        height: 2.25rem;
        padding: 0 1rem;
        background-color: var(--button-bg);
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 1rem;
        font-weight: 600;
        letter-spacing: 0.5px;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.2s ease;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .process-button:hover:not(:disabled) {
        background-color: var(--button-hover);
        transform: translateY(-1px);
    }

    .process-button:active:not(:disabled) {
        transform: scale(0.98) translateY(0);
    }

    .process-button:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    .spinner {
        width: 20px;
        height: 20px;
        border: 3px solid rgba(255, 255, 255, 0.3);
        border-radius: 50%;
        border-top-color: white;
        animation: spin 1s linear infinite;
        display: inline-block;
        margin-left: 0.5rem;
        flex-shrink: 0; /* Prevent spinner from being squished */
        line-height: 0; /* Ensure the space doesn't affect height */
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    /* Improve number input arrows */
    input[type="number"]::-webkit-inner-spin-button,
    input[type="number"]::-webkit-outer-spin-button {
        opacity: 1;
        height: 1.5em;
        margin: 0 0.25em;
    }

    /* Add subtle hover effect to inputs */
    input[type="number"]:hover {
        border-color: var(--button-bg);
    }
    
    .version {
        position: fixed;
        top: 0.2rem;
        right: 1.7rem;
        z-index: 0;
        padding: 0rem;
        font-size: 0.6rem;
        color: gray;
    }
</style>