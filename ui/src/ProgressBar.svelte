<script>
    import { Progressbar } from 'flowbite-svelte';
    import { sineOut } from 'svelte/easing';
    
    export let progress = 0;
    export let current = 0;
    export let total = 0;
    export let speed = '';
    export let currentFile = '';
    
    $: percentage = Math.min(Math.round(progress), 100).toString();

    function formatFileName(path) {
        const parts = path.split('/');
        return parts[parts.length - 1];
    }
</script>

<div class="progress-container">
    <div class="info-section">
        <div class="file-info">
            <span class="filename" title={currentFile}>
                Downloading {formatFileName(currentFile)}
            </span>
            <span class="counter">
                ({current}/{total})
            </span>
        </div>
        <div class="stats">
            <span class="percentage">
                {percentage}%
            </span>
            <span class="speed">
                {speed}
            </span>
        </div>
    </div>
    <Progressbar
        progress={percentage}
        animate={true}
        precision={0}
        size="h-2"
        color="green"
        tweenDuration={300}
        easing={sineOut}
        class="custom-progress"
    />
</div>

<style>
    .progress-container {
        background: #2a2a2a;
        border-radius: 6px;
        padding: 10px;
        margin: 4px 0;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        transition: transform 0.2s ease, box-shadow 0.2s ease;
    }

    .info-section {
        display: flex;
        justify-content: space-between;
        margin-bottom: 8px;
        font-size: 12px;
        align-items: center;
    }

    .file-info {
        flex: 1;
        min-width: 0;
        display: flex;
        gap: 8px;
        align-items: center;
    }

    .filename {
        color: #d4d4d4;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        transition: color 0.2s ease;
        padding: 2px 4px;
        border-radius: 3px;
    }

    .filename:hover {
        color: #ffffff;
        background: rgba(255, 255, 255, 0.1);
    }

    .counter {
        color: #7cafc2;
        white-space: nowrap;
        font-family: monospace;
        background: rgba(124, 175, 194, 0.1);
        padding: 2px 6px;
        border-radius: 3px;
        border: 1px solid rgba(124, 175, 194, 0.2);
    }

    .stats {
        display: flex;
        gap: 12px;
        white-space: nowrap;
        align-items: center;
    }

    .percentage {
        color: #99c794;
        min-width: 45px;
        text-align: right;
        font-weight: 600;
        font-family: monospace;
        background: rgba(153, 199, 148, 0.1);
        padding: 2px 6px;
        border-radius: 3px;
        border: 1px solid rgba(153, 199, 148, 0.2);
    }

    .speed {
        color: #7cafc2;
        min-width: 70px;
        text-align: right;
        font-family: monospace;
        background: rgba(124, 175, 194, 0.1);
        padding: 2px 6px;
        border-radius: 3px;
        border: 1px solid rgba(124, 175, 194, 0.2);
    }

    /* Custom styling for the progress bar */
    :global(.custom-progress) {
        @apply bg-gray-700;
        border-radius: 4px;
        overflow: hidden;
    }
    
    :global(.custom-progress div) {
        @apply bg-[#99c794];
        box-shadow: 0 0 10px rgba(153, 199, 148, 0.3);
        transition: all 0.3s ease;
    }

    /* Responsive design */
    @media (max-width: 480px) {
        .info-section {
            flex-direction: column;
            align-items: flex-start;
            gap: 4px;
        }

        .stats {
            width: 100%;
            justify-content: flex-end;
        }
    }

    /* High contrast focus indicators for accessibility */
    .filename:focus-visible {
        outline: 2px solid #99c794;
        outline-offset: 2px;
    }

    /* Optional: Add loading pulse effect when progress is active */
    .progress-container:not([data-complete="true"]) {
        animation: pulse 2s infinite;
    }

    @keyframes pulse {
        0% {
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        }
        50% {
            box-shadow: 0 2px 4px rgba(153, 199, 148, 0.2);
        }
        100% {
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        }
    }
</style>