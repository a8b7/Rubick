export function showToast(message: string, type: 'success' | 'error' | 'warning' | 'info') {
  const toast = (window as unknown as { toast?: { [key: string]: (msg: string) => void } }).toast
  if (toast && toast[type]) {
    toast[type](message)
  } else {
    console.log(`[${type.toUpperCase()}] ${message}`)
  }
}
