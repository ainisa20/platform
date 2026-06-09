import { pinyin } from 'pinyin-pro'

export function initialsUpper(name: string): string {
  if (!name) return ''
  return pinyin(name, { pattern: 'first', toneType: 'none', type: 'array' })
    .join('')
    .replace(/[^a-zA-Z0-9]/g, '')
    .toUpperCase()
}

export function initialsLower(name: string): string {
  return initialsUpper(name).toLowerCase()
}
