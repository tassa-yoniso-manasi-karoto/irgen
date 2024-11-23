<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

  let logs = [];
  let scrollContainer;
  let autoScroll = true;
  let isUserScrolling = false;
  
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
  
  onMount(() => {
    EventsOn("log", (log) => {
      logs = [...logs, log];
      if (autoScroll && !isUserScrolling && scrollContainer) {
        setTimeout(() => {
          scrollContainer.scrollTop = scrollContainer.scrollHeight;
        }, 0);
      }
    });
  });

  onDestroy(() => {
    EventsOff("log");
  });

  function handleScroll(e) {
    const target = e.currentTarget;
    const isAtBottom = Math.abs(
      target.scrollHeight - target.clientHeight - target.scrollTop
    ) < 1;
    
    isUserScrolling = !isAtBottom;
    if (isAtBottom) {
      autoScroll = true;
    }
  }

  function toggleAutoScroll() {
    autoScroll = !autoScroll;
    if (autoScroll && scrollContainer) {
      scrollContainer.scrollTop = scrollContainer.scrollHeight;
    }
  }
</script>

<div class="log-viewer">
  <div class="controls">
    <div class="auto-scroll">
      <input type="checkbox" id="auto-scroll" bind:checked={autoScroll} on:change={toggleAutoScroll}>
      <label for="auto-scroll">Auto-scroll</label>
    </div>
    <button on:click={() => logs = []}>Clear</button>
  </div>
  
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

  .log-container {
    flex: 1;
    overflow-y: auto;
    padding: 0;
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

  /* Updated log level colors with better contrast */
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
</style>