import puppeteer from 'puppeteer-core';
import { mkdirSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const outDir = join(__dirname, '..', 'temp');

async function main() {
  mkdirSync(outDir, { recursive: true });

  const browser = await puppeteer.launch({
    headless: true,
    executablePath: 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900 });

  console.log('Navigating to http://localhost:5173/ ...');
  await page.goto('http://localhost:5173/', { waitUntil: 'networkidle2', timeout: 30000 });
  await new Promise((r) => setTimeout(r, 2000));

  await page.screenshot({ path: join(outDir, 'screenshot-full.png'), fullPage: false });
  console.log('Full screenshot saved');

  const node = await page.$('.react-flow__node');
  if (node) {
    await node.click();
    await new Promise((r) => setTimeout(r, 500));
    await page.screenshot({ path: join(outDir, 'screenshot-node-selected.png'), fullPage: false });
    console.log('Node selected screenshot saved');
  }

  const validateBtn = await page.$('button');
  const buttons = await page.$$('button');
  for (const btn of buttons) {
    const text = await btn.evaluate((el) => el.textContent);
    if (text?.includes('验证')) {
      await btn.click();
      await new Promise((r) => setTimeout(r, 500));
      await page.screenshot({ path: join(outDir, 'screenshot-validate.png'), fullPage: false });
      console.log('Validate screenshot saved');
      break;
    }
  }

  await browser.close();
}

main().catch((err) => {
  console.error('Screenshot failed:', err);
  process.exit(1);
});
