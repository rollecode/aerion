<script lang="ts">
  import Icon from '@iconify/svelte'
  import type { Toast } from '$lib/stores/toast'
  import { _ } from '$lib/i18n'
  import { getCompactToasts } from '$lib/stores/settings.svelte'

  interface Props {
    toast: Toast
    onClose: () => void
  }

  let { toast, onClose }: Props = $props()

  const compact = $derived(getCompactToasts())

  const icons = {
    success: 'mdi:check-circle',
    error: 'mdi:alert-circle',
    info: 'mdi:information',
    warning: 'mdi:alert'
  }

  const colors = {
    success: 'bg-green-100 dark:bg-green-950 border-green-500',
    error: 'bg-red-100 dark:bg-red-950 border-red-500',
    info: 'bg-blue-100 dark:bg-blue-950 border-blue-500',
    warning: 'bg-yellow-100 dark:bg-yellow-950 border-yellow-500'
  }

  const iconColors = {
    success: 'text-green-600 dark:text-green-500',
    error: 'text-red-600 dark:text-red-500',
    info: 'text-blue-600 dark:text-blue-500',
    warning: 'text-yellow-600 dark:text-yellow-500'
  }
</script>

<div
  class="flex {compact ? 'items-center gap-2 px-3 py-1.5 rounded-md shadow-sm' : 'items-start gap-3 p-4 rounded-lg shadow-lg'} border backdrop-blur-sm {colors[toast.type]} animate-slide-in"
  role="alert"
>
  <Icon icon={icons[toast.type]} class="{compact ? 'w-4 h-4' : 'w-5 h-5 mt-0.5'} flex-shrink-0 {iconColors[toast.type]}" />

  {#if compact}
    <p class="flex-1 min-w-0 text-xs text-foreground truncate">{toast.message}</p>
    {#if toast.actions && toast.actions.length > 0}
      {#each toast.actions as action (action.label)}
        <button
          class="text-xs font-medium px-1.5 py-0.5 rounded hover:bg-white/10 transition-colors flex-shrink-0"
          onclick={action.onClick}
        >
          {action.label}
        </button>
      {/each}
    {/if}
  {:else}
    <div class="flex-1 min-w-0">
      <p class="text-sm text-foreground">{toast.message}</p>
      {#if toast.actions && toast.actions.length > 0}
        <div class="flex gap-2 mt-2">
          {#each toast.actions as action (action.label)}
            <button
              class="text-xs font-medium px-2 py-1 rounded hover:bg-white/10 transition-colors"
              onclick={action.onClick}
            >
              {action.label}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <button
    class="p-1 rounded hover:bg-white/10 transition-colors flex-shrink-0"
    onclick={onClose}
    aria-label={$_('aria.dismiss')}
  >
    <Icon icon="mdi:close" class="{compact ? 'w-3.5 h-3.5' : 'w-4 h-4'} text-muted-foreground" />
  </button>
</div>

<style>
  @keyframes slide-in {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .animate-slide-in {
    animation: slide-in 0.2s ease-out;
  }
</style>
