/**
 * Base64 编码解码工具
 */

/**
 * 编码字符串为 Base64
 * @param {string} text - 要编码的文本（支持中文）
 * @returns {string} Base64 编码后的字符串
 */
export default function encodeBase64(text) {
  try {
    const encoder = new TextEncoder();
    const bytes = encoder.encode(text);
    let binary = '';
    for (let i = 0; i < bytes.length; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  } catch (error) {
    console.error('Base64 编码失败:', error);
    throw error;
  }
}

/**
 * 解码 Base64 为字符串
 * @param {string} base64 - Base64 编码的字符串
 * @returns {string} 解码后的文本（支持中文）
 */
export function decodeBase64(base64) {
  try {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    const decoder = new TextDecoder('utf-8');
    return decoder.decode(bytes);
  } catch (error) {
    console.error('Base64 解码失败:', error);
    throw error;
  }
}

/**
 * 编码为 URL 安全的 Base64
 * 将 + 替换为 -，/ 替换为 _，并移除末尾的 =
 * @param {string} text - 要编码的文本（支持中文）
 * @returns {string} URL 安全的 Base64 编码字符串
 */
export function encodeBase64Url(text) {
  try {
    const base64 = encodeBase64(text);
    return base64
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');
  } catch (error) {
    console.error('URL 安全 Base64 编码失败:', error);
    throw error;
  }
}
