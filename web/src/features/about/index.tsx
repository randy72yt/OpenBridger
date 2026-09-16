/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { PlatformContact } from '@/components/platform-contact'
import { RichContent } from '@/components/rich-content'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { isHttpUrl, isLikelyHtml } from '@/lib/content-format'

import { getAboutContent } from './api'

export function DefaultAboutContent() {
  const { t } = useTranslation()
  const { systemName } = useSystemConfig()

  return (
    <main className='mx-auto max-w-5xl px-5 pt-28 pb-16 md:pt-32'>
      <Card className='gap-0 rounded-3xl py-0 shadow-none'>
        <CardHeader className='gap-4 p-7 md:p-12'>
          <div className='text-primary flex items-center gap-3 text-xs font-semibold tracking-wider'>
            {systemName} · {t('Regional access notice')}
          </div>
          <h1 className='text-3xl font-semibold tracking-tight md:text-4xl'>
            {t('Service usage notice')}
          </h1>
          <p className='text-muted-foreground leading-8'>
            {t(
              'This notice explains service eligibility, permitted use, account responsibilities, operational logs and abuse handling. Please read it before using the platform.'
            )}
          </p>
        </CardHeader>
        <CardContent className='space-y-8 px-7 pb-7 md:px-12 md:pb-12'>
          <Card className='bg-primary/5 ring-primary/15 rounded-2xl py-0 shadow-none'>
            <CardContent className='space-y-3 p-6'>
              <h2 className='text-lg font-semibold'>{t('Important notice')}</h2>
              <p className='text-muted-foreground leading-8'>
                {t(
                  'This platform serves only users outside mainland China. Please ensure that your access to and use of the service comply with applicable laws in your location, upstream provider terms and platform rules. Responsibility for violations rests with the respective responsible parties in accordance with the law.'
                )}
              </p>
            </CardContent>
          </Card>
          <Separator className='mx-auto max-w-[min(36rem,80%)] opacity-60' />
          <section className='space-y-3'>
            <h2 className='text-lg font-semibold'>{t('Permitted use')}</h2>
            <p className='text-muted-foreground leading-8'>
              {t(
                'Use your account and API keys only for your own projects or authorized team activities. Do not sell or publicly distribute credentials, resell access without authorization, bypass limits, attack systems or infringe others’ rights. Follow applicable laws, provider terms and academic integrity rules.'
              )}
            </p>
          </section>
          <Separator className='mx-auto max-w-[min(36rem,80%)] opacity-60' />
          <section className='space-y-3'>
            <h2 className='text-lg font-semibold'>{t('Data and security')}</h2>
            <p className='text-muted-foreground leading-8'>
              {t(
                'To deliver requests, measure usage and investigate issues, the platform processes necessary request data and may record login IPs, request times, models, usage, status and error information. Requests are forwarded to the selected upstream provider. Avoid submitting unnecessary sensitive information; consult the privacy policy for data handling details.'
              )}
            </p>
          </section>
          <Separator className='mx-auto max-w-[min(36rem,80%)] opacity-60' />
          <section className='space-y-3'>
            <h2 className='text-lg font-semibold'>{t('Account security')}</h2>
            <p className='text-muted-foreground leading-8'>
              {t(
                'Keep your password and API keys private. Check usage regularly, and revoke any key that may have been exposed. Do not include credentials in support messages.'
              )}
            </p>
          </section>
          <Separator className='mx-auto max-w-[min(36rem,80%)] opacity-60' />
          <section className='space-y-3'>
            <h2 className='text-lg font-semibold'>{t('Abuse handling')}</h2>
            <p className='text-muted-foreground leading-8'>
              {t(
                'Abnormal access, API abuse, attempts to bypass restrictions or other security risks may lead to restrictions, suspension or termination of affected accounts and keys. Measures depend on the risk; urgent security incidents may require action without advance notice. Contact the platform if you believe a restriction is mistaken.'
              )}
            </p>
          </section>
          <Separator className='mx-auto max-w-[min(36rem,80%)] opacity-60' />
          <section id='business' className='scroll-mt-24 space-y-3'>
            <h2 className='text-lg font-semibold'>
              {t('Business cooperation')}
            </h2>
            <p className='text-muted-foreground leading-8'>
              {t(
                'Discuss API integration, team access or model service partnerships with us. Include your use case, expected usage and technical requirements when contacting us.'
              )}
            </p>
          </section>
          <section id='contact' className='scroll-mt-24 space-y-3'>
            <h2 className='text-lg font-semibold'>
              {t('footer.columns.about.links.contact')}
            </h2>
            <div className='bg-muted/60 text-muted-foreground flex flex-col gap-3 rounded-2xl p-6 text-sm'>
              <PlatformContact />
            </div>
          </section>
        </CardContent>
      </Card>
    </main>
  )
}

export function About() {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: ['about-content'],
    queryFn: getAboutContent,
  })

  const rawContent = data?.data?.trim() ?? ''
  const hasContent = rawContent.length > 0
  const isUrl = hasContent && isHttpUrl(rawContent)
  const contentIsHtml = hasContent && isLikelyHtml(rawContent)

  if (isLoading) {
    return (
      <PublicLayout>
        <div className='mx-auto flex max-w-4xl flex-col gap-4 py-12'>
          <Skeleton className='h-8 w-[45%]' />
          <Skeleton className='h-4 w-full' />
          <Skeleton className='h-4 w-[90%]' />
          <Skeleton className='h-4 w-[80%]' />
        </div>
      </PublicLayout>
    )
  }

  if (!hasContent) {
    return (
      <PublicLayout showMainContainer={false}>
        <DefaultAboutContent />
        <Footer />
      </PublicLayout>
    )
  }

  if (isUrl) {
    return (
      <PublicLayout showMainContainer={false}>
        <iframe
          src={rawContent}
          className='h-[calc(100vh-3.5rem)] w-full border-0'
          title={t('About')}
          sandbox='allow-forms allow-popups allow-popups-to-escape-sandbox allow-scripts'
        />
      </PublicLayout>
    )
  }

  if (contentIsHtml) {
    return (
      <PublicLayout showMainContainer={false}>
        <RichContent
          mode='html'
          htmlVariant='isolated'
          content={rawContent}
          className='prose-neutral dark:prose-invert max-w-none'
        />
      </PublicLayout>
    )
  }

  return (
    <PublicLayout>
      <div className='mx-auto max-w-6xl px-4 py-8'>
        <RichContent
          mode='markdown'
          content={rawContent}
          className='prose-neutral dark:prose-invert max-w-none'
        />
      </div>
    </PublicLayout>
  )
}
