import DOMPurify from 'dompurify'

/**
 * Sanitize HTML content to prevent XSS attacks.
 * Allows safe formatting tags (bold, italic, links, lists, etc.)
 * but strips dangerous elements like <script>, event handlers, etc.
 */
export function sanitizeHTML(html) {
  if (!html) return ''
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'p', 'br', 'hr',
      'strong', 'em', 'b', 'i', 'u', 's', 'strike', 'del', 'ins',
      'a', 'img',
      'ul', 'ol', 'li',
      'blockquote', 'pre', 'code',
      'table', 'thead', 'tbody', 'tr', 'th', 'td',
      'div', 'span',
      'sub', 'sup',
    ],
    ALLOWED_ATTR: [
      'href', 'target', 'rel',
      'src', 'alt', 'width', 'height',
      'title', 'class', 'id',
      'style',
    ],
    ALLOW_DATA_ATTR: false,
  })
}
