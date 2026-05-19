const CHARS = 'abcdefghijklmnopqrstuvwxyz0123456789';

export function generateNodeId(existingIds?: string[]): string {
  let id = '';
  for (let attempt = 0; attempt < 100; attempt++) {
    id = '';
    for (let i = 0; i < 4; i++) {
      id += CHARS[Math.floor(Math.random() * CHARS.length)];
    }
    if (!existingIds || !existingIds.includes(id)) {
      return id;
    }
  }
  return id + Date.now().toString(36).slice(-4);
}
