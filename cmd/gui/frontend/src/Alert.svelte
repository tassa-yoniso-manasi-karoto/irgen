<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Alert } from 'flowbite-svelte';
    import { 
        InfoCircle,
        ExclamationCircle,
        CheckCircle,
        ExclamationTriangle
    } from 'svelte-bootstrap-icons';
    
    const dispatch = createEventDispatcher();
    
    export let message: string;
    export let visible: boolean = true;
    export let type: 'debug' | 'info' | 'warn' | 'error' | 'fatal' | 'panic' | 'trace' = 'error';
    
    function handleDismiss() {
        visible = false;
        dispatch('dismiss');
    }

    $: IconComponent = {
        debug: InfoCircle,
        info: InfoCircle,
        warn: ExclamationTriangle,
        error: ExclamationCircle,
        fatal: ExclamationCircle,
        panic: ExclamationCircle,
        trace: InfoCircle
    }[type];
</script>


{#if visible}
    <div class="alert-overlay">
        <Alert
            class="custom-alert {type}"
            dismissable
            on:dismiss={handleDismiss}
        >
            <svelte:fragment slot="icon">
                <svelte:component this={IconComponent} width="48" height="48" />
            </svelte:fragment>
            <span class="font-medium">
                {message}
            </span>
        </Alert>
    </div>
{/if}

<style>
    .alert-overlay {
        position: fixed;
        top: 1rem;
        left: 50%;
        transform: translateX(-50%);
        z-index: 1000;
        width: max-content;
        max-width: calc(100vw - 2rem);
        animation: slideDown 0.3s ease-out;
    }

    @keyframes slideDown {
        from {
            transform: translate(-50%, -100%);
            opacity: 0;
        }
        to {
            transform: translate(-50%, 0);
            opacity: 1;
        }
    }

    :global(.custom-alert) {
        background-color: #1e1e1e !important;
        border: 1px solid;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
        transition: all 0.2s ease-in-out;
        backdrop-filter: blur(8px);
    }
    
    :global(.custom-alert:hover) {
        transform: translateY(-1px);
        box-shadow: 0 6px 16px rgba(0, 0, 0, 0.6);
    }
    
    :global(.custom-alert.debug) {
        border-color: #7cafc2;
        color: #7cafc2 !important;
        background-color: #1a2329 !important;
    }
    
    :global(.custom-alert.info) {
        border-color: #99c794;
        color: #99c794 !important;
        background-color: #1a291e !important;
    }
    
    :global(.custom-alert.warn) {
        border-color: #fac863;
        color: #fac863 !important;
        background-color: #292419 !important;
    }
    
    :global(.custom-alert.error) {
        border-color: #ec5f67;
        color: #ec5f67 !important;
        background-color: #291a1a !important;
    }
    
    :global(.custom-alert.fatal) {
        border-color: #ff8080;
        color: #ff8080 !important;
        background-color: #291616 !important;
    }
    
    :global(.custom-alert.panic) {
        border-color: #ff6b6b;
        color: #ff6b6b !important;
        background-color: #291515 !important;
    }
    
    :global(.custom-alert.trace) {
        border-color: #c792ea;
        color: #c792ea !important;
        background-color: #231929 !important;
    }

    :global(.custom-alert) {
        padding: 0.75rem 1.25rem !important;
        display: flex;
        align-items: center;
        gap: 0.75rem;
        min-width: 300px;
    }

    :global(.custom-alert button) {
        color: currentColor !important;
        opacity: 0.8;
        padding: 0.25rem;
        border-radius: 4px;
        transition: all 0.2s ease-in-out;
    }

    :global(.custom-alert button:hover) {
        opacity: 1;
        background-color: rgba(255, 255, 255, 0.1);
    }
    
    :global(.custom-alert svg) {
        flex-shrink: 0;
    }

    :global(.custom-alert span) {
        line-height: 1.5;
        font-size: 0.95rem;
    }
</style>