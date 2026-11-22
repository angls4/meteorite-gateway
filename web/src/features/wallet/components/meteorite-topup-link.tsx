/*
meteorite: the only top-up change of the gateway (D-069). The test top-up is a page of the
meteorite service on this same host. It is offered only while the service says it is turned on.
The sign-in goes along in the address fragment, which browsers never send to a server, so the
page can prove to the gateway whose balance to add to.
*/
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { buttonVariants } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'

export function MeteoriteTopupLink() {
  const { t } = useTranslation()
  const [enabled, setEnabled] = useState(false)
  const auth = useAuthStore((state) => state.auth)

  useEffect(() => {
    let current = true
    fetch('/topup/api/status')
      .then((response) => (response.ok ? response.json() : { enabled: false }))
      .then((body: { enabled?: boolean }) => {
        if (current) setEnabled(body.enabled === true)
      })
      .catch(() => {
        if (current) setEnabled(false)
      })
    return () => {
      current = false
    }
  }, [])

  if (!enabled || !auth.accessToken || !auth.user) return null
  const fragment = new URLSearchParams({
    t: auth.accessToken,
    u: String(auth.user.id),
  })
  return (
    <div className='flex flex-wrap items-center justify-between gap-3 rounded-lg border px-4 py-3'>
      <p className='text-muted-foreground text-sm'>
        {t('Test top-up adds balance without a payment. It is for development only.')}
      </p>
      <a href={`/topup#${fragment.toString()}`} className={buttonVariants({ variant: 'outline' })}>
        {t('Add test balance')}
      </a>
    </div>
  )
}
