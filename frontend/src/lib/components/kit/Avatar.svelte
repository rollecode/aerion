<script module lang="ts">
  const bimiCache = new Map<string, string>()

  const TWO_PART_TLDS = new Set(['co.uk','com.au','co.jp','co.nz','com.br','com.mx','com.cn','co.in','net.au','org.uk','me.uk','ac.uk','co.kr','com.sg','com.hk','com.tw','com.ar','com.co','com.pe','co.za','com.ua','co.il','com.tr','com.pl','com.eg','com.sa','com.ae','com.ph','com.my','com.vn','com.th'])

  function apexDomain(domain: string): string {
    const parts = domain.split('.')
    if (parts.length <= 2) return domain
    const lastTwo = parts.slice(-2).join('.')
    return TWO_PART_TLDS.has(lastTwo) ? parts.slice(-3).join('.') : lastTwo
  }
</script>

<script lang="ts">
  import { md5 } from '$lib/utils/md5'

  interface Props {
    email: string
    name?: string
    density?: 'micro' | 'compact' | 'standard' | 'large'
    size?: number
    photoData?: string
    photoMediaType?: string
    gravatar?: boolean
    /** Draw the coloured initials circle when no image resolves. Off leaves the
     *  slot empty so the layout still lines up. */
    initialsFallback?: boolean
  }

  const { email, name, density = 'standard', size, photoData, photoMediaType, gravatar = true, initialsFallback = true }: Props = $props()

  let photoFailed = $state(false)
  let gravatarFailed = $state(false)
  let faviconFailed = $state(false)
  let appleTouchIconFailed = $state(false)
  let dicebearFailed = $state(false)
  let bimiFailed = $state(false)
  let bimiUrl = $state('')

  function colorClass(seed: string): string {
    let hash = 0
    for (let i = 0; i < seed.length; i++) {
      hash = seed.charCodeAt(i) + ((hash << 5) - hash)
    }
    return `avatar-${(Math.abs(hash) % 14) + 1}`
  }

  function initials(displayName: string | undefined, fallbackEmail: string): string {
    if (!displayName && !fallbackEmail) return '?'
    const name = displayName || fallbackEmail
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const DENSITY_SIZE: Record<NonNullable<Props['density']>, number> = {
    micro: 24,
    compact: 28,
    standard: 32,
    large: 40,
  }

  const px = $derived(size ?? DENSITY_SIZE[density])
  const fontPx = $derived(Math.round(px * 0.4))
  const cls = $derived(colorClass(email || ''))
  const text = $derived(initials(name, email))

  const normalizedEmail = $derived(email?.trim().toLowerCase() ?? '')
  const domain = $derived(normalizedEmail.split('@')[1] ?? '')
  const apex = $derived(apexDomain(domain))

  const gravatarUrl = $derived(
    gravatar && normalizedEmail && !gravatarFailed
      ? `https://www.gravatar.com/avatar/${md5(normalizedEmail)}?d=404&s=${Math.round(px * 2)}`
      : '',
  )

  const appleTouchIconUrl = $derived(
    gravatar && apex && !appleTouchIconFailed
      ? `https://${apex}/apple-touch-icon.png`
      : '',
  )

  const faviconUrl = $derived(
    gravatar && apex && !faviconFailed
      ? `https://t1.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&fallback_opts=TYPE,SIZE,URL&url=https://${apex}&size=128`
      : '',
  )

  const dicebearUrl = $derived(
    gravatar && normalizedEmail && !dicebearFailed
      ? `https://api.dicebear.com/9.x/notionists-neutral/png?seed=${encodeURIComponent(normalizedEmail)}&size=${Math.round(px * 2)}`
      : '',
  )

  const showBimi = $derived(!!bimiUrl && !bimiFailed)
  const showGravatar = $derived(!showBimi && !!gravatarUrl)
  const showPhoto = $derived(!showBimi && !showGravatar && !!photoData && !!photoMediaType && !photoFailed)
  const showAppleTouchIcon = $derived(!showBimi && !showGravatar && !showPhoto && !!appleTouchIconUrl)
  const showFavicon = $derived(!showBimi && !showGravatar && !showPhoto && !showAppleTouchIcon && !!faviconUrl)
  const showDicebear = $derived(!showBimi && !showGravatar && !showPhoto && !showAppleTouchIcon && !showFavicon && !!dicebearUrl)

  $effect(() => {
    if (!gravatar || !apex) {
      bimiUrl = ''
      return
    }
    const cached = bimiCache.get(apex)
    if (cached !== undefined) {
      bimiUrl = cached
      return
    }
    bimiUrl = ''
    fetch(`https://dns.google/resolve?name=default._bimi.${encodeURIComponent(apex)}&type=TXT`)
      .then((r) => r.json())
      .then((j) => {
        const record = j?.Answer?.[0]?.data ?? ''
        const logo = record.split(';').map((s: string) => s.trim()).find((s: string) => s.startsWith('l='))?.slice(2) ?? ''
        bimiCache.set(apex, logo)
        bimiUrl = logo
      })
      .catch(() => {
        bimiCache.set(apex, '')
        bimiUrl = ''
      })
  })
</script>

<div
  class="rounded-full flex-shrink-0 inline-flex items-center justify-center font-medium overflow-hidden {showBimi || showGravatar || showPhoto || showAppleTouchIcon || showFavicon || showDicebear || !initialsFallback ? '' : cls}"
  style:width="{px}px"
  style:height="{px}px"
  style:font-size="{fontPx}px"
  aria-hidden="true"
>
  {#if showBimi}
    <img
      src={bimiUrl}
      alt=""
      class="w-full h-full object-cover"
      onerror={() => { bimiFailed = true }}
    />
  {:else if showGravatar}
    <img
      src={gravatarUrl}
      alt=""
      class="w-full h-full object-cover"
      onerror={() => { gravatarFailed = true }}
    />
  {:else if showPhoto}
    <img
      src="data:{photoMediaType};base64,{photoData}"
      alt=""
      class="w-full h-full object-cover"
      onerror={() => { photoFailed = true }}
    />
  {:else if showAppleTouchIcon}
    <img
      src={appleTouchIconUrl}
      alt=""
      class="w-full h-full object-cover"
      onerror={() => { appleTouchIconFailed = true }}
    />
  {:else if showFavicon}
    <img
      src={faviconUrl}
      alt=""
      class="w-full h-full object-cover"
      onload={(e) => { if ((e.currentTarget as HTMLImageElement).naturalWidth <= 16) faviconFailed = true }}
      onerror={() => { faviconFailed = true }}
    />
  {:else if showDicebear}
    <img
      src={dicebearUrl}
      alt=""
      class="w-full h-full object-cover"
      onerror={() => { dicebearFailed = true }}
    />
  {:else if initialsFallback}
    {text}
  {/if}
</div>
