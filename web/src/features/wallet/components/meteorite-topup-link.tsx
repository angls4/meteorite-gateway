/*
meteorite: the links of this gateway to pages of the meteorite service on this same host (D-069,
R-017): the test top-up, offered only while the service says it is turned on, and the download of
the user's own data. The sign-in goes along in the address fragment, which browsers never send to
a server, so the page can prove to the gateway which account it acts on.
*/
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { buttonVariants } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'

type MeteoriteStatus = { enabled: boolean; account: boolean }

export function MeteoriteTopupLink() {
  const { t } = useTranslation()
  const [status, setStatus] = useState<MeteoriteStatus>({
    enabled: false,
    account: false,
  })
  const auth = useAuthStore((state) => state.auth)

  useEffect(() => {
    let current = true
    fetch('/topup/api/status')
      .then((response) => (response.ok ? response.json() : null))
      .then((body: Partial<MeteoriteStatus> | null) => {
        if (current && body)
          setStatus({ enabled: body.enabled === true, account: body.account === true })
      })
      .catch(() => undefined)
    return () => {
      current = false
    }
  }, [])

  if ((!status.enabled && !status.account) || !auth.accessToken || !auth.user)
    return null
  const fragment = new URLSearchParams({
    t: auth.accessToken,
    u: String(auth.user.id),
  }).toString()
  return (
    <div className='flex flex-col gap-3 rounded-lg border px-4 py-3'>
      {status.enabled && (
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <p className='text-muted-foreground text-sm'>
            {t('Test top-up adds balance without a payment. It is for development only.')}
          </p>
          <a href={`/topup#${fragment}`} className={buttonVariants({ variant: 'outline' })}>
            {t('Add test balance')}
          </a>
        </div>
      )}
      {status.account && (
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <p className='text-muted-foreground text-sm'>
            {t('Download your account data: keys (masked), request metadata of the last 90 days and balance transactions.')}
          </p>
          <a href={`/account#${fragment}`} className={buttonVariants({ variant: 'outline' })}>
            {t('Download my data')}
          </a>
        </div>
      )}
    </div>
  )
}
