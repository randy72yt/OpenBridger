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
import { zodResolver } from '@hookform/resolvers/zod'
import { ImagePlus, Trash2 } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { useRef, type ChangeEvent } from 'react'
import type { Resolver } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import * as z from 'zod'

import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { DEFAULT_SUPPORT_EMAIL, isQrImageSource } from '@/lib/platform-contact'

import { FormDirtyIndicator } from '../components/form-dirty-indicator'
import { FormNavigationGuard } from '../components/form-navigation-guard'
import {
  SettingsForm,
  SettingsFormGrid,
  SettingsFormGridItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useSettingsForm } from '../hooks/use-settings-form'
import { useUpdateOption } from '../hooks/use-update-option'
import { isValidTaskPublicAddress } from './task-public-address'

const MAX_QR_FILE_BYTES = 5 * 1024 * 1024
const MAX_QR_DIMENSION = 420

const _systemInfoSchema = z.object({
  SystemName: z.string().min(1),
  ServerAddress: z.string().optional(),
  TaskPublicAddress: z.string().refine(isValidTaskPublicAddress),
  Logo: z.string().url().optional().or(z.literal('')),
  Footer: z.string().optional(),
  About: z.string().optional(),
  HomePageContent: z.string().optional(),
  legal: z.object({
    user_agreement: z.string().optional(),
    privacy_policy: z.string().optional(),
  }),
  contact: z.object({
    email: z.string().optional(),
    wechat_id: z.string().optional(),
    wechat_qrcode: z.string().optional(),
  }),
  landing: z.object({
    featured_models: z.string().max(200).optional(),
  }),
})

type SystemInfoFormValues = z.infer<typeof _systemInfoSchema>

type SystemInfoSectionProps = {
  defaultValues: SystemInfoFormValues
}

function normalizeValue(value: unknown): string {
  if (value === undefined || value === null) return ''
  return typeof value === 'string' ? value : String(value)
}

function compressQrImageToDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.addEventListener(
      'error',
      () => reject(new Error('Failed to read image')),
      { once: true }
    )
    reader.addEventListener(
      'load',
      () => {
        const dataUrl = String(reader.result ?? '')
        const img = new Image()
        img.addEventListener(
          'error',
          () => reject(new Error('Invalid image file')),
          { once: true }
        )
        img.addEventListener(
          'load',
          () => {
            const scale = Math.min(
              1,
              MAX_QR_DIMENSION / Math.max(img.width || 1, img.height || 1)
            )
            const targetWidth = Math.max(1, Math.round(img.width * scale))
            const targetHeight = Math.max(1, Math.round(img.height * scale))
            const canvas = document.createElement('canvas')
            canvas.width = targetWidth
            canvas.height = targetHeight
            const ctx = canvas.getContext('2d')
            if (!ctx) {
              resolve(dataUrl)
              return
            }
            ctx.fillStyle = '#FFFFFF'
            ctx.fillRect(0, 0, targetWidth, targetHeight)
            ctx.drawImage(img, 0, 0, targetWidth, targetHeight)
            const webp = canvas.toDataURL('image/webp', 0.88)
            if (webp.startsWith('data:image/webp')) {
              resolve(webp)
              return
            }
            resolve(canvas.toDataURL('image/png'))
          },
          { once: true }
        )
        img.src = dataUrl
      },
      { once: true }
    )
    reader.readAsDataURL(file)
  })
}

export function SystemInfoSection(props: SystemInfoSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const qrFileInputRef = useRef<HTMLInputElement>(null)
  const defaultValues = props.defaultValues

  const normalizedDefaults: SystemInfoFormValues = {
    SystemName: normalizeValue(defaultValues.SystemName),
    ServerAddress: normalizeValue(defaultValues.ServerAddress),
    TaskPublicAddress: normalizeValue(defaultValues.TaskPublicAddress),
    Logo: normalizeValue(defaultValues.Logo),
    Footer: normalizeValue(defaultValues.Footer),
    About: normalizeValue(defaultValues.About),
    HomePageContent: normalizeValue(defaultValues.HomePageContent),
    legal: {
      user_agreement: normalizeValue(defaultValues.legal?.user_agreement),
      privacy_policy: normalizeValue(defaultValues.legal?.privacy_policy),
    },
    contact: {
      email: normalizeValue(defaultValues.contact?.email),
      wechat_id: normalizeValue(defaultValues.contact?.wechat_id),
      wechat_qrcode: normalizeValue(defaultValues.contact?.wechat_qrcode),
    },
    landing: {
      featured_models: normalizeValue(defaultValues.landing?.featured_models),
    },
  }

  const systemInfoSchemaWithI18n = z.object({
    SystemName: z.string().min(1, {
      error: () => t('System name is required'),
    }),
    ServerAddress: z.string().optional(),
    TaskPublicAddress: z.string().refine(isValidTaskPublicAddress, {
      error: () =>
        t(
          'Enter an absolute HTTP(S) URL without credentials, query parameters, or fragments'
        ),
    }),
    Logo: z.string().url().optional().or(z.literal('')),
    Footer: z.string().optional(),
    About: z.string().optional(),
    HomePageContent: z.string().optional(),
    legal: z.object({
      user_agreement: z.string().optional(),
      privacy_policy: z.string().optional(),
    }),
    contact: z.object({
      email: z.string().email().optional().or(z.literal('')),
      wechat_id: z.string().max(128).optional(),
      wechat_qrcode: z.string().optional(),
    }),
    landing: z.object({
      featured_models: z.string().max(200).optional(),
    }),
  })

  const { form, handleSubmit, handleReset, isDirty, isSubmitting } =
    useSettingsForm<SystemInfoFormValues>({
      resolver: zodResolver(systemInfoSchemaWithI18n) as Resolver<
        SystemInfoFormValues,
        unknown,
        SystemInfoFormValues
      >,
      defaultValues: normalizedDefaults,
      onSubmit: async (_data, changedFields) => {
        for (const [key, value] of Object.entries(changedFields)) {
          let v = normalizeValue(value)
          if (key === 'ServerAddress' || key === 'TaskPublicAddress') {
            v = v.replace(/\/+$/, '')
          }
          await updateOption.mutateAsync({
            key,
            value: v,
          })
        }
      },
    })

  const handleQrFileSelected = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return

    if (!/^image\/(png|jpe?g|webp|gif)$/i.test(file.type)) {
      toast.error(t('Please upload a PNG, JPG, WebP, or GIF image'))
      return
    }
    if (file.size > MAX_QR_FILE_BYTES) {
      toast.error(t('Image file must be smaller than 5 MB'))
      return
    }

    try {
      const compressedDataUrl = await compressQrImageToDataUrl(file)
      form.setValue('contact.wechat_qrcode', compressedDataUrl, {
        shouldDirty: true,
        shouldValidate: true,
      })
    } catch {
      toast.error(t('Failed to process QR code image'))
    }
  }

  return (
    <>
      <FormNavigationGuard when={isDirty} />

      <SettingsSection title={t('System Information')}>
        <Form {...form}>
          <SettingsForm onSubmit={handleSubmit}>
            <SettingsPageFormActions
              onSave={handleSubmit}
              onReset={handleReset}
              isSaving={isSubmitting || updateOption.isPending}
              isResetDisabled={!isDirty}
            />
            <FormDirtyIndicator isDirty={isDirty} />
            <SettingsFormGrid>
              <FormField
                control={form.control}
                name='SystemName'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('System Name')}</FormLabel>
                    <FormControl>
                      <Input placeholder={t('OpenBridger')} {...field} />
                    </FormControl>
                    <FormDescription>
                      {t('The name displayed across the application')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='ServerAddress'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Server Address')}</FormLabel>
                    <FormControl>
                      <Input placeholder='https://yourdomain.com' {...field} />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'The public URL of your server, used for OAuth callbacks, webhooks, and other external integrations'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='TaskPublicAddress'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Async Task Public Address')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='https://media.example.com/tasks'
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Public base URL for async task media. Supports a dedicated media domain, port, or Nginx path prefix; falls back to Server Address when empty.'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='Logo'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Logo URL')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('https://example.com/logo.png')}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('URL to your logo image (optional)')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='contact.email'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Support & Business Email')}</FormLabel>
                    <FormControl>
                      <Input placeholder={DEFAULT_SUPPORT_EMAIL} {...field} />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Displayed on the Partner Program and About pages for business and support inquiries'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='contact.wechat_id'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Business WeChat ID')}</FormLabel>
                    <FormControl>
                      <Input placeholder='OpenBridger_AI' {...field} />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'WeChat ID displayed on the Partner Program and About pages with one-click copy'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <SettingsFormGridItem span='full'>
                <FormField
                  control={form.control}
                  name='contact.wechat_qrcode'
                  render={({ field }) => {
                    const qrValue = (field.value ?? '').trim()
                    const isDataUrl = qrValue.startsWith('data:image/')
                    const isImage = isQrImageSource(qrValue)

                    return (
                      <FormItem>
                        <FormLabel>{t('Business WeChat QR Code')}</FormLabel>
                        <div className='grid gap-4 sm:grid-cols-[1fr_auto] sm:items-start'>
                          <div className='space-y-2.5'>
                            <FormControl>
                              <Input
                                placeholder='https://example.com/wechat-qr.png'
                                value={
                                  isDataUrl
                                    ? t('(Uploaded QR image)')
                                    : (field.value ?? '')
                                }
                                readOnly={isDataUrl}
                                onChange={field.onChange}
                              />
                            </FormControl>
                            <div className='flex flex-wrap items-center gap-2'>
                              <input
                                ref={qrFileInputRef}
                                type='file'
                                accept='image/png,image/jpeg,image/webp,image/gif'
                                className='hidden'
                                onChange={handleQrFileSelected}
                              />
                              <Button
                                type='button'
                                variant='outline'
                                size='sm'
                                onClick={() => qrFileInputRef.current?.click()}
                              >
                                <ImagePlus className='mr-1.5 size-4' />
                                {t('Upload QR image')}
                              </Button>
                              {qrValue && (
                                <Button
                                  type='button'
                                  variant='ghost'
                                  size='sm'
                                  className='text-muted-foreground hover:text-destructive'
                                  onClick={() =>
                                    form.setValue('contact.wechat_qrcode', '', {
                                      shouldDirty: true,
                                      shouldValidate: true,
                                    })
                                  }
                                >
                                  <Trash2 className='mr-1.5 size-4' />
                                  {t('Clear')}
                                </Button>
                              )}
                            </div>
                            <FormDescription>
                              {t(
                                'Upload a QR screenshot directly or paste an image URL / WeChat QR link. Changes take effect immediately without redeployment.'
                              )}
                            </FormDescription>
                            <FormMessage />
                          </div>

                          {qrValue && (
                            <div className='flex flex-col items-center gap-1.5 rounded-xl border bg-white p-2.5 shadow-2xs'>
                              {isImage ? (
                                <img
                                  src={qrValue}
                                  alt={t('Business WeChat QR Code')}
                                  className='size-28 rounded-md object-contain'
                                />
                              ) : (
                                <QRCodeSVG
                                  value={qrValue}
                                  size={112}
                                  level='M'
                                />
                              )}
                            </div>
                          )}
                        </div>
                      </FormItem>
                    )
                  }}
                />
              </SettingsFormGridItem>

              <FormField
                control={form.control}
                name='Footer'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Footer')}</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={t(
                          '© 2025 Your Company. All rights reserved.'
                        )}
                        rows={4}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Footer text displayed at the bottom of pages')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='About'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('About')}</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={t(
                          'Enter HTML code (e.g., <p>About us...</p>) or a URL (e.g., https://example.com) to embed as iframe'
                        )}
                        rows={4}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Supports HTML markup or iframe embedding. Enter HTML code directly, or provide a complete URL to automatically embed it as an iframe.'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <SettingsFormGridItem span='full'>
                <FormField
                  control={form.control}
                  name='landing.featured_models'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t('Homepage Featured Models Tagline')}
                      </FormLabel>
                      <FormControl>
                        <Input
                          placeholder='GPT-6.2 · Codex · Claude 5.5 · Gemini 4 · DeepSeek V4.1'
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        {t(
                          'Optional tagline shown in the homepage hero metrics. Leave empty to automatically showcase the latest models from your active channels.'
                        )}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </SettingsFormGridItem>

              <SettingsFormGridItem span='full'>
                <FormField
                  control={form.control}
                  name='HomePageContent'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Home Page Content')}</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder={t('Welcome to our OpenBridger...')}
                          rows={6}
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        {t(
                          'Content displayed on the home page (supports Markdown)'
                        )}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </SettingsFormGridItem>

              <FormField
                control={form.control}
                name='legal.user_agreement'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('User Agreement')}</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={t(
                          'Provide Markdown, HTML, or an external URL for the user agreement'
                        )}
                        rows={6}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Leave empty to disable the agreement requirement. Supports Markdown, HTML, or a full URL to redirect users.'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='legal.privacy_policy'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Privacy Policy')}</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={t(
                          'Provide Markdown, HTML, or an external URL for the privacy policy'
                        )}
                        rows={6}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Leave empty to disable the privacy policy requirement. Supports Markdown, HTML, or a full URL to redirect users.'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </SettingsFormGrid>
          </SettingsForm>
        </Form>
      </SettingsSection>
    </>
  )
}
