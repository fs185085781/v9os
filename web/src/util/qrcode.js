import QRCode from 'qrcode';

/**
 * 生成二维码 Data URL
 * @param {string} text - 要编码的文本内容
 * @param {object} options - 二维码配置选项
 * @returns {Promise<string>} Data URL
 */
export async function generateQrCode(text, options = {}) {
  const defaultOptions = {
    width: 200,
    margin: 1,
    color: {
      dark: '#000000',
      light: '#ffffff',
    },
    errorCorrectionLevel: 'M',
    ...options,
  };

  try {
    const dataUrl = await QRCode.toDataURL(text, defaultOptions);
    return dataUrl;
  } catch (error) {
    console.error('生成二维码失败:', error);
    throw error;
  }
}

/**
 * 生成二维码 Canvas（用于更复杂的场景）
 * @param {HTMLCanvasElement} canvas - Canvas 元素
 * @param {string} text - 要编码的文本内容
 * @param {object} options - 二维码配置选项
 * @returns {Promise<void>}
 */
export async function generateQrCodeToCanvas(canvas, text, options = {}) {
  const defaultOptions = {
    width: 200,
    margin: 1,
    color: {
      dark: '#000000',
      light: '#ffffff',
    },
    errorCorrectionLevel: 'M',
    ...options,
  };

  try {
    await QRCode.toCanvas(canvas, text, defaultOptions);
  } catch (error) {
    console.error('生成二维码到Canvas失败:', error);
    throw error;
  }
}
