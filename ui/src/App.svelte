<script lang="ts">
    import { onMount } from 'svelte';
    import LogViewer from './LogViewer.svelte';
    import ThemeToggle from './ThemeToggle.svelte';
    
    let url = '';
    let numberOfTitle = 3;
    let maxXResolution = 1920;
    let maxYResolution = 1080;
    let status = '';
    let isProcessing = false;
    let isDark = true; // Default to dark mode
    let version = "0.0.0";
    
    
    onMount(async () => {
        version = await window.go.gui.App.GetVersion();
    });
    
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
</script>
<div class="theme-toggle-wrapper">
    <span class="version">irgen {version}</span>
    <ThemeToggle bind:isDark />
</div>

<div class="app-background" class:dark={isDark}>
    <main class="container">
        <div class="url-input">
            <input
                type="text"
                bind:value={url}
                placeholder="Enter URL or path to HTML file here"
            />
        </div>

        <div class="number-inputs-wrapper">
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
        </div>

        <div class="button-container">
            <button on:click={processURL} disabled={isProcessing}>
                {#if isProcessing}
                    Processing... 
                    <div class="spinner"></div>
                {:else}
                    Process
                {/if}
            </button>
        </div>
        
        <LogViewer />

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

    .url-input {
        margin-bottom: 1rem;
    }

    .url-input input {
        width: 99%;
        padding: 0.75rem;
        border: 1px solid var(--input-border);
        border-radius: 4px;
        font-size: 1rem;
        box-sizing: border-box;
        background-color: var(--input-bg);
        color: var(--text-color);
    }

    .url-input input::placeholder {
        color: var(--text-color);
        opacity: 0.6;
    }

    .number-inputs-wrapper {
        width: 100%;
        margin-bottom: 1rem;
        overflow-x: auto;
    }

    .number-inputs {
        display: flex;
        justify-content: space-between;
        gap: 1rem;
        min-width: 600px;
    }

    .input-group {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        min-width: 180px;
    }

    label {
        margin-bottom: 0.5rem;
        font-weight: bold;
        color: var(--text-color);
        white-space: nowrap;
        text-align: center;
    }

    input[type="number"] {
        width: 100%;
        padding: 0.5rem;
        border: 1px solid var(--input-border);
        border-radius: 4px;
        text-align: center;
        font-size: 1rem;
        box-sizing: border-box;
        background-color: var(--input-bg);
        color: var(--text-color);
    }

    .button-container {
        display: flex;
        justify-content: center;
        margin-bottom: 1rem;
    }

    button {
        width: auto;
        min-width: 200px;
        padding: 0.75rem 1.5rem;
        background-color: var(--button-bg);
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 1.1rem;
        font-weight: bold;
        transition: all 0.3s ease;
    }

    button:hover {
        background-color: var(--button-hover);
    }

    button:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    .status {
        padding: 1rem;
        background-color: var(--container-bg);
        border-radius: 4px;
        white-space: pre-line;
        margin-top: 1rem;
        color: var(--text-color);
    }
    
    .spinner {
        width: 20px;
        height: 20px;
        border: 3px solid rgba(255, 255, 255, 0.3);
        border-radius: 50%;
        border-top-color: white;
        animation: spin 1s linear infinite;
        display: inline-block;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    button:active:not(:disabled) {
        transform: scale(0.98);
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